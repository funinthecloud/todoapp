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
	"github.com/funinthecloud/protosource/stores/dynamodbstore"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	client := dynamodb.NewFromConfig(cfg)

	eventsTable := dynamodbstore.EventsTableName(envOrDefault("EVENTS_TABLE", "events"))
	aggregatesTable := dynamodbstore.AggregatesTableName(envOrDefault("AGGREGATES_TABLE", "aggregates"))

	authorizer, err := InitializeAuthorizer(client, eventsTable, aggregatesTable)
	if err != nil {
		panic(err)
	}

	router, err := InitializeRouter(client, eventsTable, aggregatesTable)
	if err != nil {
		panic(err)
	}

	router.Handle("GET", "whoami", whoamiHandler(authorizer))

	inner := awslambda.WrapRouter(router, extractActor)
	lambda.Start(corsWrapper(inner))
}

// corsWrapper adds CORS headers with credentials support to every Lambda
// response. Validates the request Origin against CORS_ALLOWED_ORIGINS
// (comma-separated) or defaults to https://todoapp.drhayt.com.
func corsWrapper(
	next func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error),
) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	allowed := buildAllowedOrigins()
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		origin := req.Headers["origin"]
		if origin == "" {
			origin = req.Headers["Origin"]
		}

		if !allowed[origin] {
			origin = ""
		}

		if req.HTTPMethod == http.MethodOptions && origin != "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNoContent,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      origin,
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTIONS",
					"Access-Control-Allow-Headers":     "Content-Type,X-Actor,Authorization",
					"Vary":                             "Origin",
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
			resp.Headers["Vary"] = "Origin"
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

func buildAllowedOrigins() map[string]bool {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "https://todoapp.drhayt.com"
	}
	m := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			m[o] = true
		}
	}
	return m
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
