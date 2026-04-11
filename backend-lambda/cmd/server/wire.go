//go:build wireinject

package main

import (
	"os"

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

	"github.com/funinthecloud/protosource-auth/authz/httpauthz"

	todolistv1 "github.com/funinthecloud/todoapp/backend-lambda/gen/showcase/app/todolist/v1"
	todolistv1dynamodb "github.com/funinthecloud/todoapp/backend-lambda/gen/showcase/app/todolist/v1/todolistv1dynamodb"
)

func provideRouter(
	todolistHandler *todolistv1.Handler,
) *protosource.Router {
	return protosource.NewRouter(todolistHandler)
}

// provideAuthorizer returns an httpauthz.Authorizer pointing at the
// protosource-auth service URL supplied by PROTOSOURCE_AUTH_URL, or
// allowall.Authorizer{} when the env var is empty. Production lambda
// deployments set PROTOSOURCE_AUTH_URL to the auth service's endpoint.
func provideAuthorizer() authz.Authorizer {
	if url := os.Getenv("PROTOSOURCE_AUTH_URL"); url != "" {
		return httpauthz.New(url)
	}
	return allowall.Authorizer{}
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
		dynamodbstore.ProviderSet,
		protobinaryserializer.ProviderSet,
		todolistv1dynamodb.ProviderSet,
		todolistv1.NewTodoListClient,
		provideAuthorizer,
		todolistv1.NewHandler,
		provideRouter,
	)
	return nil, nil
}
