//go:build wireinject

package main

import (
	"github.com/funinthecloud/protosource"
	todolistv1 "github.com/funinthecloud/protosource-showcase/backend/gen/showcase/app/todolist/v1"
	"github.com/funinthecloud/protosource/serializers/protobinaryserializer"
	"github.com/funinthecloud/protosource/stores/memorystore"
	"github.com/google/wire"
)

func provideStore() *memorystore.MemoryStore {
	return memorystore.New(todolistv1.SnapshotEveryNEvents)
}

func provideRepository(store *memorystore.MemoryStore, serializer *protobinaryserializer.Serializer) *protosource.Repository {
	return todolistv1.NewRepository(store, serializer)
}

func provideHandler(repo *protosource.Repository) *todolistv1.Handler {
	return todolistv1.NewHandler(repo, nil)
}

func InitializeRepository() *protosource.Repository {
	wire.Build(
		provideStore,
		protobinaryserializer.ProviderSet,
		provideRepository,
	)
	return nil
}

func InitializeHandler(repo *protosource.Repository) *todolistv1.Handler {
	wire.Build(
		provideHandler,
	)
	return nil
}
