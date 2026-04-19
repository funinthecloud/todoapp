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

	authorizer := InitializeAuthorizer()

	router, err := InitializeRouter(client, eventsTable, aggregatesTable)
	if err != nil {
		panic(err)
	}

	router.Handle("GET", "whoami", whoamiHandler(authorizer))

	router.SetCORS(protosource.CORSConfig{
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,X-Actor,Authorization",
	})

	handler := awslambda.WrapRouter(router, extractActor)
	lambda.Start(handler)
}

func whoamiHandler(authorizer authz.Authorizer) protosource.HandlerFunc {
	return func(ctx context.Context, req protosource.Request) protosource.Response {
		enrichedCtx, err := authorizer.Authorize(ctx, req, "")
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
		body, _ := json.Marshal(map[string]string{"actor": actor})
		return protosource.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}
	}
}

// extractActor prefers an Authorization: Bearer <shadow-token> header
// (dereferenced by the httpauthz Authorizer wired in wire.go) and
// falls back to X-Actor for allowall/developer mode. Returns
// "anonymous" when neither header is present so the generated
// handler's CMD_NO_ACTOR check still passes.
func extractActor(req events.APIGatewayProxyRequest) string {
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

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
