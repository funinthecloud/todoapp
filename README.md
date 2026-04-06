# todoapp

A showcase app for [protosource](https://github.com/funinthecloud/protosource) -- event sourcing where domain models are defined entirely in protocol buffers.

This is a to-do list manager with two interchangeable backends and a React frontend.

## Prerequisites

- Go 1.25+
- Node 18+
- [buf](https://buf.build/docs/installation)
- `protoc-gen-go` and `protoc-gen-protosource` installed to `$GOPATH/bin`:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install github.com/funinthecloud/protosource/cmd/protoc-gen-protosource@latest
  ```
- [wire](https://github.com/goforj/wire) installed to `$GOPATH/bin`:
  ```bash
  go install github.com/goforj/wire/cmd/wire@latest
  ```

## Project Structure

```
proto/                          # Domain model (shared by all backends)
backend-bolt/                   # BoltDB backend (local dev, zero cloud deps)
backend-lambda/                 # DynamoDB + AWS Lambda backend
frontend/                       # React + Vite + TypeScript
```

---

## Option A: Local Development (backend-bolt)

BoltDB stores everything in a local file. No AWS account, no Docker, no database to install.

### 1. Generate code and start the backend

```bash
buf generate
cd backend-bolt
go mod tidy
wire ./cmd/server/
go run ./cmd/server/
```

The server listens on `http://localhost:8080`. Data is persisted to `backend-bolt/data/`.

### 2. Start the frontend

```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8080 npm run dev
```

Open `http://localhost:5173`.

---

## Option B: AWS Deployment (backend-lambda)

DynamoDB for persistence, Lambda for compute. Deploys behind API Gateway.

### 1. Generate code

```bash
buf generate --template buf.gen.lambda.yaml
cd backend-lambda
go mod tidy
wire ./cmd/server/
```

### 2. Create DynamoDB tables

```bash
go run ./cmd/setup/ create
```

This creates two tables (`events` and `aggregates`) with:
- PAY_PER_REQUEST billing
- 20 GSIs on the aggregates table
- TTL and PITR enabled
- Deletion protection enabled

Override table names with `EVENTS_TABLE` and `AGGREGATES_TABLE` env vars.

Use `go run ./cmd/setup/ status` to verify, or `go run ./cmd/setup/ fix` to enable TTL/PITR on existing tables.

### 3. Build the Lambda binary

```bash
GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/server/
zip function.zip bootstrap
```

### 4. Deploy

Deploy `function.zip` as an ARM64 Lambda behind API Gateway (HTTP API, proxy integration) using your preferred tool (SAM, CDK, Terraform, console).

Lambda environment variables:
| Variable | Default | Description |
|---|---|---|
| `EVENTS_TABLE` | `events` | DynamoDB events table name |
| `AGGREGATES_TABLE` | `aggregates` | DynamoDB aggregates table name |

### 5. Start the frontend

Point the frontend at your API Gateway URL:

```bash
cd frontend
npm install
VITE_API_URL=https://abc123.execute-api.us-east-1.amazonaws.com npm run dev
```

For a production build:

```bash
VITE_API_URL=https://abc123.execute-api.us-east-1.amazonaws.com npm run build
```

Static files are output to `frontend/dist/`.

---

## Domain Model

A single `TodoList` aggregate with a `map<string, TodoItem>` collection.

**Commands:** Create, Rename, Archive, Unarchive, AddItem, UpdateItem, RemoveItem

**States:** Active, Archived

**Derived fields** (`AfterOn` hook): `item_count`, `completed_count` -- computed from the items collection after every event replay.

The proto definition is in `proto/showcase/app/todolist/v1/todolist_v1.proto`.

---

## Modifying the Proto

After changing `todolist_v1.proto`:

```bash
clang-format --style=file -i proto/**/*.proto
buf generate
buf generate --template buf.gen.lambda.yaml
cd backend-bolt && wire ./cmd/server/
cd ../backend-lambda && wire ./cmd/server/
```
