# CLAUDE.md

Showcase app for protosource: a to-do list manager demonstrating event sourcing with protocol buffers.

## Build & Run

```bash
# Backend (from repo root)
cd backend
go generate ./...                            # install tools (wire)
go install github.com/funinthecloud/protosource/cmd/protoc-gen-protosource@latest  # install plugin
buf generate                                 # generate Go code from proto/ (run from repo root)
wire ./cmd/server/                           # generate wire_gen.go
go build ./cmd/server/                       # build server
go run ./cmd/server/                         # run on :8080

# Frontend
cd frontend
npm install
npm run dev                                  # Vite dev server on :5173
```

After modifying proto files:
```bash
buf generate                                 # from repo root
cd backend && wire ./cmd/server/             # regenerate wire if providers changed
```

## Architecture

- **Proto** (`proto/showcase/app/todolist/v1/`) -- domain model definition
- **Backend** (`backend/`) -- Go server using protosource framework with memorystore (in-memory, no DB needed)
- **Frontend** (`frontend/`) -- React + Vite + TypeScript, uses `@protosource/client`
- **Generated** (`backend/gen/`) -- buf-generated Go code (gitignored, regenerate with `buf generate`)

Hand-written file in gen directory:
- `backend/gen/showcase/app/todolist/v1/todolist_derived.go` -- `AfterOn()` hook for derived fields (item_count, completed_count)

## Proto Formatting

Proto files MUST be formatted with `clang-format`:
```bash
clang-format --style=file -i proto/**/*.proto
```

## Domain

Single aggregate `TodoList` with `map<string, TodoItem>` collection. Commands: Create, Rename, Archive, Unarchive, AddItem, UpdateItem, RemoveItem. UpdateItem replaces the full item (use for toggling completed, editing title, etc).
