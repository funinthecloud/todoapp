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

# Deploy via your preferred method (SAM, CDK, Terraform, etc.)
```

Environment variables for Lambda: `EVENTS_TABLE` (default: events), `AGGREGATES_TABLE` (default: aggregates).

### Frontend

```bash
cd frontend
npm install
npm run dev                                      # Vite dev server on :5173
```

### After modifying proto files

```bash
clang-format --style=file -i proto/**/*.proto    # format first
buf generate                                     # bolt backend
buf generate --template buf.gen.lambda.yaml      # lambda backend
cd backend-bolt && wire ./cmd/server/            # regenerate wire if needed
cd ../backend-lambda && wire ./cmd/server/
```

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
