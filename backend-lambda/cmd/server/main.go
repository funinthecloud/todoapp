package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/awslambda"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/authz/allowall"
	"github.com/funinthecloud/protosource/stores/dynamodbstore"
	"github.com/funinthecloud/protosource-auth/authz/httpauthz"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	client := dynamodb.NewFromConfig(cfg)

	eventsTable := dynamodbstore.EventsTableName(envOrDefault("EVENTS_TABLE", "events"))
	aggregatesTable := dynamodbstore.AggregatesTableName(envOrDefault("AGGREGATES_TABLE", "aggregates"))

	authorizer := provideAuthorizer()

	router, err := InitializeRouter(client, eventsTable, aggregatesTable, authorizer)
	if err != nil {
		panic(err)
	}

	router.Handle("GET", "whoami", whoamiHandler(authorizer))

	inner := awslambda.WrapRouter(router, extractActor)
	lambda.Start(corsWrapper(inner))
}

// provideAuthorizer returns httpauthz when PROTOSOURCE_AUTH_URL is set,
// otherwise falls back to allowall for local development.
func provideAuthorizer() authz.Authorizer {
	authURL := os.Getenv("PROTOSOURCE_AUTH_URL")
	if authURL == "" {
		return allowall.Authorizer{}
	}
	return httpauthz.New(authURL, httpauthz.WithTokenSource(
		httpauthz.Chain(httpauthz.Cookie("shadow"), httpauthz.AuthorizationHeader()),
	))
}

// corsWrapper adds CORS headers with credentials support to every Lambda
// response. Echoes the request Origin header and handles OPTIONS preflight.
func corsWrapper(
	next func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error),
) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		origin := req.Headers["origin"]
		if origin == "" {
			origin = req.Headers["Origin"]
		}

		if req.HTTPMethod == http.MethodOptions && origin != "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNoContent,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      origin,
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTIONS",
					"Access-Control-Allow-Headers":     "Content-Type,X-Actor,Authorization",
				},
			}, nil
		}

		resp, err := next(ctx, req)
		if origin != "" {
			if resp.Headers == nil {
				resp.Headers = make(map[string]string)
			}
			resp.Headers["Access-Control-Allow-Origin"] = origin
			resp.Headers["Access-Control-Allow-Credentials"] = "true"
		}
		return resp, err
	}
}

func whoamiHandler(authorizer authz.Authorizer) protosource.HandlerFunc {
	return func(ctx context.Context, req protosource.Request) protosource.Response {
		enrichedCtx, err := authorizer.Authorize(ctx, req, "showcase.app.todolist.v1.WhoAmI")
		if err != nil {
			body, _ := json.Marshal(map[string]string{"error": "unauthorized"})
			return protosource.Response{
				StatusCode: http.StatusUnauthorized,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(body),
			}
		}
		actor := authz.UserIDFromContext(enrichedCtx)
		if actor == "" {
			actor = req.Actor
		}
		if actor == "" || actor == "anonymous" {
			body, _ := json.Marshal(map[string]string{"error": "unauthorized"})
			return protosource.Response{
				StatusCode: http.StatusUnauthorized,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(body),
			}
		}
		body, _ := json.Marshal(map[string]string{"actor": actor})
		return protosource.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}
	}
}

// extractActor prefers a "shadow" cookie (HttpOnly, set by the auth
// service), then falls back to Authorization: Bearer <token> and X-Actor
// for developer convenience. Returns "anonymous" when no identity can be
// determined so the generated handler's CMD_NO_ACTOR check still passes.
func extractActor(req events.APIGatewayProxyRequest) string {
	for _, key := range []string{"cookie", "Cookie"} {
		if cookieHeader := req.Headers[key]; cookieHeader != "" {
			if v := parseCookieValue(cookieHeader, "shadow"); v != "" {
				return v
			}
		}
	}
	for _, key := range []string{"Authorization", "authorization"} {
		if auth := req.Headers[key]; strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if actor := req.Headers["x-actor"]; actor != "" {
		return actor
	}
	if actor := req.Headers["X-Actor"]; actor != "" {
		return actor
	}
	return "anonymous"
}

// parseCookieValue extracts a named cookie value from a raw Cookie header.
func parseCookieValue(header, name string) string {
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if eqIdx := strings.IndexByte(part, '='); eqIdx > 0 {
			if part[:eqIdx] == name {
				return part[eqIdx+1:]
			}
		}
	}
	return ""
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
