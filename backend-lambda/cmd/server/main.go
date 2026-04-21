package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/awslambda"
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

	router, err := InitializeRouter(client, eventsTable, aggregatesTable)
	if err != nil {
		panic(err)
	}

	router.SetCORS(protosource.CORSConfig{
		AllowOrigins:     buildAllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,X-Actor,Authorization",
		AllowCredentials: true,
	})

	lambda.Start(awslambda.WrapRouter(router, extractActor))
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

func buildAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "https://todoapp.drhayt.com"
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
