# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

- Module: `github.com/dangngochoainam/codebase-go-grpc` · Go 1.24
- gRPC v1.76 + grpc-gateway v2.27 · GORM v1.31 (PostgreSQL) · Uber Dig v1.19 · Zap v1.27
- RabbitMQ (amqp091 v1.11) · protovalidate v1.1

## Commands

```bash
# Docker dev environment
make start              # docker compose -f docker-compose.debug.yaml up -d
make attach             # Shell into the running debug container
make mod                # go mod tidy && go mod vendor

# Run
make run                # go run cmd/server/main.go
make run-watch          # Live reload via CompileDaemon (Delve on port 2302)

# Code generation (re-run when .proto files change)
make gen-proto          # Regenerate pb/ from proto/
# Note: gen-proto uses ${GOPATH} glob for grpc-gateway include.
# If it fails, run protoc directly with the explicit path:
# protoc --proto_path=proto -I $(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@<version> \
#   --go_out=pb --go_opt=paths=source_relative \
#   --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
#   --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
#   proto/*.proto

# Testing
go test ./...
go vet ./...
```

## Architecture

Clean architecture with 4 layers, wired via Uber Dig dependency injection (`internal/diregistry/`):

```
HTTP/gRPC → API (internal/api/) → Use Case (internal/usecase/) → Repository (internal/repository/) → DB (GORM + PostgreSQL)
```

**Dual protocol**: gRPC on port 50050, HTTP REST via gRPC-Gateway on port 50051. Both started in `cmd/server/main.go`.

**Middleware chain** (applied to every request in order):
1. Trace ID injection
2. Protobuf validation (`buf.build/protovalidate`)
3. Request/response payload logging (zap)
4. Panic recovery

### Layer responsibilities

| Layer | Location | Role |
|-------|----------|------|
| Entry point | `cmd/server/main.go` | Build DI container, start gRPC + HTTP servers, graceful shutdown |
| API | `internal/api/` | gRPC handler implementations; converts protobuf ↔ DTOs |
| Use Case | `internal/usecase/` | Business logic; orchestrates repositories |
| Repository | `internal/repository/` | Data access via GORM |
| Entities | `db/entities/` | GORM model definitions; drive schema migrations |

### Key packages

Shared helpers live in the external [`gopkg`](https://github.com/dangngochoainam/gopkg) repo (mounted locally via `replace` in `go.mod`). Refer to that repo for up-to-date package docs.

### Protobuf → Go workflow

1. Edit `.proto` files under `proto/`
2. Run `make gen-proto` → regenerates `pb/`
3. Implement or update handlers in `internal/api/`

### Database workflow (Atlas + GORM)

Schema is derived from GORM entity structs (`db/entities/`): **Product**, **Job** (both embed **BaseModel** with audit fields and soft delete).

1. Add/modify GORM entity structs in `db/entities/`
2. Register new entities in `db/golang-migrate/loader/main.go`
3. Generate migration: `cd db/golang-migrate && ./shmake diff <migration_name>`
4. Apply migration: `./shmake apply`

### Dependency injection

All dependencies registered in `internal/diregistry/diregistry.go` using Uber Dig in this order: Config → ModelConverter + PbConverter + HttpClientHelper → GormHelper → MinioStorage → ProductRepository + JobRepository → ProductUseCase + ExportUseCase + JobUseCase → API. When adding a new service/repository/use case, register its constructor there and add it as a parameter to the consuming constructor.

**Pattern: always inject via constructor parameters — never instantiate directly inside a constructor.**

```go
// WRONG — hard-coded, bypasses DI
func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{converter: copyhelper.NewModelConverter()}
}

// CORRECT — receive as a parameter so Dig injects it
func NewProductRepository(db *gorm.DB, converter copyhelper.ModelConverter) ProductRepository {
    return &productRepository{db: db, converter: converter}
}
```

### Copying between structs

Use the project's copier helpers when copying structs into the **output/result** returned from a function. The rule does **not** apply to constructing input structs passed into the next layer — manual field assignment is correct there.

| Helper | Method | When to use |
|--------|--------|-------------|
| `PbConverter` | `FromPb(dst, src)` | Protobuf request → DTO — at the **head** of an API handler |
| `PbConverter` | `ToPb(dst, src)` | DTO → Protobuf response — at the **end** of an API handler |
| `ModelConverter` | `FromModel(dst, src)` | DB entity/repo output → DTO — at the **end** of a use case / repository function |
| `ModelConverter` | `ToModel(dst, src)` | DTO → DB entity — at the **end** of a repository function |

```go
// WRONG — manual field assignment when copying into a return value
reqDTO.Limit = request.GetLimit()
reqDTO.After = request.GetAfter()

// CORRECT — use the converter at the head of the API handler
a.pbConverter.FromPb(reqDTO, request)
```

Fields that cannot be mapped by name (e.g. pointer unwrapping, type mismatches) are the only valid reason to assign manually on a return-value copy, and only for those specific fields after the converter call.

## Configuration

Copy `.env.example` to `.env`. Viper uses `__` as the key separator, so nested config fields map to env vars like:

| Config field | Env var |
|---|---|
| `database_postgres.host` | `DATABASE_POSTGRES__HOST` |
| `grpc_port` / `http_port` | `GRPC_PORT` / `HTTP_PORT` |

Default config is embedded in `configs/config.go` (used when no env overrides exist). Defaults: gRPC `50050`, HTTP `50051`, PostgreSQL `localhost:5434`.

## Docker

- `docker-compose.debug.yaml` + `Dockerfile.debug` — development environment with Delve debugger (port 2302) and CompileDaemon hot reload; ports `50060:50050` (gRPC), `50061:50051` (HTTP), `2302:2302` (Delve); mounts entire repo into `/app`

## Working Principles

### 1. Think Before Coding

Don't assume. Don't hide confusion. Surface tradeoffs.

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

Minimum code that solves the problem. Nothing speculative.

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

Touch only what you must. Clean up only your own mess.

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it — don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

Define success criteria. Loop until verified.

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
