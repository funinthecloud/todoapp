//go:build wireinject

package main

import (
	"os"

	"github.com/goforj/wire"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/authz/allowall"
	"github.com/funinthecloud/protosource/serializers/protobinaryserializer"
	"github.com/funinthecloud/protosource/stores/boltdbstore"

	"github.com/funinthecloud/protosource-auth/authz/httpauthz"

	todolistv1 "github.com/funinthecloud/todoapp/backend-bolt/gen/showcase/app/todolist/v1"
)

func provideStore() (*boltdbstore.BoltDBStore, error) {
	return boltdbstore.New("data", "todolist")
}

func provideRepository(store *boltdbstore.BoltDBStore, serializer *protobinaryserializer.Serializer) *protosource.Repository {
	return todolistv1.NewRepository(store, serializer)
}

// provideAuthorizer returns an httpauthz.Authorizer pointing at the
// protosource-auth service URL supplied by PROTOSOURCE_AUTH_URL, or
// falls back to allowall.Authorizer{} for local development without
// a running auth service. Any caller that wants a hard dependency on
// the auth service should set the env var — allowall is purely a
// convenience for "run backend-bolt with no auth infrastructure".
func provideAuthorizer() authz.Authorizer {
	if url := os.Getenv("PROTOSOURCE_AUTH_URL"); url != "" {
		return httpauthz.New(url)
	}
	return allowall.Authorizer{}
}

func provideHandler(repo *protosource.Repository, authorizer authz.Authorizer) *todolistv1.Handler {
	return todolistv1.NewHandler(repo, nil, authorizer)
}

func InitializeRepository() (*protosource.Repository, error) {
	wire.Build(
		provideStore,
		protobinaryserializer.ProviderSet,
		provideRepository,
	)
	return nil, nil
}

func InitializeHandler(repo *protosource.Repository) *todolistv1.Handler {
	wire.Build(
		provideAuthorizer,
		provideHandler,
	)
	return nil
}
