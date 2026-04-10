# Single source of truth for the protosource version.
# Extracted from backend-bolt/go.mod so the plugins always match the library.
PROTOSOURCE_VERSION := $(shell awk '/funinthecloud\/protosource v/{print $$2; exit}' backend-bolt/go.mod)

.PHONY: tools gen gen-bolt gen-lambda gen-ts tidy

tools:
	@echo "Installing protosource plugins @ $(PROTOSOURCE_VERSION)"
	go install github.com/funinthecloud/protosource/cmd/protoc-gen-protosource@$(PROTOSOURCE_VERSION)
	go install github.com/funinthecloud/protosource/cmd/protoc-gen-protosource-ts@$(PROTOSOURCE_VERSION)

gen: tools gen-bolt gen-lambda gen-ts tidy

gen-bolt:
	clang-format --style=file -i proto/showcase/app/todolist/v1/*.proto
	buf generate

gen-lambda:
	buf generate --template buf.gen.lambda.yaml

gen-ts:
	PATH="frontend/node_modules/.bin:$$PATH" buf generate --template buf.gen.ts.yaml

tidy:
	cd backend-bolt && go mod tidy && wire ./cmd/server/
	cd backend-lambda && go mod tidy && wire ./cmd/server/
