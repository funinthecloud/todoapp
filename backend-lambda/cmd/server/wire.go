//go:build wireinject

package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/goforj/wire"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/aws/dynamoclient"
	"github.com/funinthecloud/protosource/opaquedata"
	opaquedynamo "github.com/funinthecloud/protosource/opaquedata/dynamo"
	"github.com/funinthecloud/protosource/serializers/protobinaryserializer"
	"github.com/funinthecloud/protosource/stores/dynamodbstore"

	"github.com/funinthecloud/protosource-auth/authz/directauthz"
	"github.com/funinthecloud/protosource-auth/authz/httpauthz"
	rolev1 "github.com/funinthecloud/protosource-auth/gen/auth/role/v1"
	rolev1dynamodb "github.com/funinthecloud/protosource-auth/gen/auth/role/v1/rolev1dynamodb"
	tokenv1 "github.com/funinthecloud/protosource-auth/gen/auth/token/v1"
	tokenv1dynamodb "github.com/funinthecloud/protosource-auth/gen/auth/token/v1/tokenv1dynamodb"
	userv1 "github.com/funinthecloud/protosource-auth/gen/auth/user/v1"
	userv1dynamodb "github.com/funinthecloud/protosource-auth/gen/auth/user/v1/userv1dynamodb"
	"github.com/funinthecloud/protosource-auth/service"

	todolistv1 "github.com/funinthecloud/todoapp/backend-lambda/gen/showcase/app/todolist/v1"
)

func provideRouter(
	todolistHandler *todolistv1.Handler,
) *protosource.Router {
	return protosource.NewRouter(todolistHandler)
}

func provideChecker(
	tokenRepo tokenv1.Repo,
	userRepo userv1.Repo,
	roleRepo rolev1.Repo,
) *service.Checker {
	return service.NewChecker(tokenRepo, userRepo, roleRepo)
}

func provideAuthorizer(checker *service.Checker) authz.Authorizer {
	return directauthz.New(checker, directauthz.WithTokenSource(
		httpauthz.Chain(httpauthz.Cookie("shadow"), httpauthz.AuthorizationHeader()),
	))
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
		todolistv1.ProviderSet,
		todolistv1.NewTodoListClient,
		todolistv1.NewHandler,
		provideRouter,

		// Auth: directauthz backed by shared DynamoDB tables.
		tokenv1dynamodb.ProviderSet,
		userv1dynamodb.ProviderSet,
		rolev1dynamodb.ProviderSet,
		provideChecker,
		provideAuthorizer,
	)
	return nil, nil
}

