# CLAUDE.md

Showcase app for protosource: a to-do list manager demonstrating event sourcing with protocol buffers.

## Build & Run

### backend-bolt (BoltDB, local dev -- zero infra)

```bash
buf generate                                     # from repo root
cd backend-bolt
go mod tidy
wire ./cmd/server/
go build ./cmd/server/
go run ./cmd/server/                             # listens on :8080, data persisted to data/
```

### backend-lambda (DynamoDB + AWS Lambda)

```bash
buf generate --template buf.gen.lambda.yaml      # from repo root
cd backend-lambda
go mod tidy
wire ./cmd/server/

# Create DynamoDB tables (once)
go run ./cmd/setup/ create

# Build Lambda binary
GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/server/

# Deploy via SAM
sam build && sam deploy --guided                  # first time
sam build && sam deploy                           # subsequent
```

Environment variables for Lambda: `EVENTS_TABLE` (default: events), `AGGREGATES_TABLE` (default: aggregates).

### Frontend

```bash
cd frontend
npm install
npm run dev                                      # Vite dev server on :5173
VITE_API_URL=https://todov1.api.drhayt.com npm run build  # production build
```

### After modifying proto files

```bash
make gen                                         # format, install plugins, generate all, tidy + wire
```

Individual targets: `make tools`, `make gen-bolt`, `make gen-lambda`, `make gen-ts`, `make tidy`.
`PROTOSOURCE_VERSION` is extracted from `backend-bolt/go.mod` so the plugins always match the library.

## Architecture

- **Proto** (`proto/showcase/app/todolist/v1/`) -- domain model definition
- **backend-bolt/** -- Go HTTP server using BoltDB (local file persistence, no cloud deps)
- **backend-lambda/** -- AWS Lambda handler using DynamoDB (events + aggregates tables)
  - `cmd/server/` -- Lambda entrypoint
  - `cmd/setup/` -- DynamoDB table creation CLI (create, fix, delete, status)
- **Frontend** (`frontend/`) -- React + Vite + TypeScript, uses `@protosource/client`

Each backend has its own Go module and generated code. Hand-written `todolist_derived.go` (AfterOn hook) exists in both `backend-*/gen/showcase/app/todolist/v1/`.

## Proto Formatting

```bash
clang-format --style=file -i proto/**/*.proto
```

## Domain

Single aggregate `TodoList` with `map<string, TodoItem>` collection. Commands: Create, Rename, Archive, Unarchive, AddItem, UpdateItem, RemoveItem. UpdateItem replaces the full item (use for toggling completed, editing title, etc).

## Authorization

Both backends wire an `authz.Authorizer` via `provideAuthorizer()` in `cmd/server/wire.go`:

- `PROTOSOURCE_AUTH_URL` set → `httpauthz.New(url)` dereferences each request's `Authorization: Bearer <shadow-token>` against the running auth service via HTTP
- `PROTOSOURCE_AUTH_URL` unset → `allowall.Authorizer{}` (developer flow, trusts the `X-Actor` header)
- **Planned:** switch to `directauthz.New(checker)` for the Lambda backend — in-process authorization against shared DynamoDB tables, no HTTP round-trip. See `protosource-auth/authz/directauthz/`.

The `actorExtractor` in each backend's `main.go` prefers `Authorization: Bearer <token>` and falls back to `X-Actor`. On the generated handler side, protosource v0.1.3's template reads `authz.UserIDFromContext(ctx)` first — so in httpauthz mode the aggregate's `create_by` / `modify_by` fields carry the resolved user id, not the raw bearer token.

End-to-end against DynamoDB Local: see [protosource-auth/README.md](https://github.com/funinthecloud/protosource-auth#readme) for the auth-service side, then run backend-bolt with `PROTOSOURCE_AUTH_URL=http://localhost:8080`.

## Gotchas

- API Gateway lowercases all request headers. Check `x-actor` not `X-Actor` in Lambda extractors.
- DynamoDB GSIs require both PK and SK attributes on an item to project it. If GSI SK is missing, the item silently disappears from query results.
- `protosource_opaque_field` annotations on enum fields require protosource >= v0.0.8.
- When bumping protosource past v0.1.2, also bump the BSR buf module (`buf dep update`) and run `wire` in both backends — the `NewHandler` signature gains an authorizer parameter.

## Upstream dependencies

- **`github.com/funinthecloud/protosource`** — the framework. Local at `$HOME/Developer/funinthecloud/protosource`. Codegen plugins `protoc-gen-protosource` (Go) and `protoc-gen-protosource-ts` (TS) installed by `make tools`.
- **`github.com/funinthecloud/protosource-auth`** — the shadow-token auth service. Imports `authz/httpauthz` (HTTP-based authorizer) and will migrate to `authz/directauthz` (in-process, shared DynamoDB tables) for the Lambda backend. Also provides `service.Checker` for direct authorization. Local at `$HOME/Developer/funinthecloud/protosource-auth`.

Upgrade workflow: `go get github.com/funinthecloud/protosource@vX.Y.Z` in both `backend-bolt` and `backend-lambda`, `buf dep update` at the repo root, then `make gen`. Wire regeneration happens automatically as part of `make gen`.
