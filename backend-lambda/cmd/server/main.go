package main

import (
	"context"
	"os"

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
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,X-Actor",
	})

	handler := awslambda.WrapRouter(router, extractActor)
	lambda.Start(handler)
}

func extractActor(req events.APIGatewayProxyRequest) string {
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
