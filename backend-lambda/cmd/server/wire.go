//go:build wireinject

package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/goforj/wire"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/authz/allowall"
	"github.com/funinthecloud/protosource/aws/dynamoclient"
	"github.com/funinthecloud/protosource/opaquedata"
	opaquedynamo "github.com/funinthecloud/protosource/opaquedata/dynamo"
	"github.com/funinthecloud/protosource/serializers/protobinaryserializer"
	"github.com/funinthecloud/protosource/stores/dynamodbstore"
	todolistv1 "github.com/funinthecloud/todoapp/backend-lambda/gen/showcase/app/todolist/v1"
)

func provideRouter(
	todolistHandler *todolistv1.Handler,
) *protosource.Router {
	return protosource.NewRouter(todolistHandler)
}

// InitializeRouter wires all dependencies and returns a configured router.
func InitializeRouter(
	client *dynamodb.Client,
	eventsTable dynamodbstore.EventsTableName,
	aggregatesTable dynamodbstore.AggregatesTableName,
) (*protosource.Router, error) {
	wire.Build(
		wire.Bind(new(dynamoclient.Client), new(*dynamodb.Client)),
		wire.Bind(new(opaquedata.OpaqueStore), new(*opaquedynamo.Store)),
		allowall.ProviderSet,
		dynamodbstore.ProviderSet,
		protobinaryserializer.ProviderSet,
		todolistv1.ProviderSet,
		todolistv1.NewTodoListClient,
		todolistv1.NewHandler,
		provideRouter,
	)
	return nil, nil
}

func InitializeAuthorizer() authz.Authorizer {
	wire.Build(allowall.ProviderSet)
	return nil
}
