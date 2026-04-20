//go:build wireinject

package main

import (
	"github.com/goforj/wire"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/serializers/protobinaryserializer"
	"github.com/funinthecloud/protosource/stores/boltdbstore"
	todolistv1 "github.com/funinthecloud/todoapp/backend-bolt/gen/showcase/app/todolist/v1"
)

func provideStore() (*boltdbstore.BoltDBStore, error) {
	return boltdbstore.New("data", "todolist")
}

func provideRepository(store *boltdbstore.BoltDBStore, serializer *protobinaryserializer.Serializer) *protosource.Repository {
	return todolistv1.NewRepository(store, serializer)
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

func InitializeHandler(repo *protosource.Repository, authorizer authz.Authorizer) *todolistv1.Handler {
	wire.Build(
		provideHandler,
	)
	return nil
}
