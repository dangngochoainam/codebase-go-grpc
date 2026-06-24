# codebase-go-grpc

## 1. Resources

You only need two things installed locally:

- **Docker + Docker Compose**
- **`docker_stack` external network** with the following services running on it:

| Service | Container hostname | Port |
|---|---|---|
| PostgreSQL | `postgres` | 5432 |

All Go tooling (protoc, CompileDaemon, Delve, protoc-gen-*) is pre-installed inside the container via `Dockerfile.debug` — no local setup required.

If the network doesn't exist yet:
```bash
docker network create docker_stack
```

## 2. Run with watch

```bash
make start      # build image and start the container
make attach     # open a shell inside the container

# Inside the container:
make mod        # download and vendor deps (first time, or after go.mod changes)
make run-watch  # live reload via CompileDaemon + Delve debugger
```

Ports exposed to the host:

| Protocol | Host port |
|---|---|
| gRPC | `50060` |
| HTTP REST | `50061` |
| Delve debugger | `2302` |

> First run only — apply DB migrations inside the container before starting the app:
> ```bash
> cd db/golang-migrate
> # edit export-env.local.sh with DB credentials (host: postgres, port: 5432)
> ./shmake build
> ./shmake diff init
> ./shmake apply
> ```

## 3. Update database schema (entity changes)

See [`db/golang-migrate/README.md`](db/golang-migrate/README.md) for the full workflow.

## 4. Regenerate proto

After editing any `.proto` file under `proto/`, run inside the container:

```bash
make gen-proto
```

This regenerates all files in `pb/`. Commit the updated `pb/` files alongside your `.proto` changes, then update any affected handlers in `internal/api/`.

If `make gen-proto` fails due to a GOPATH glob issue, run protoc directly:

```bash
protoc --proto_path=proto \
  -I $(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@<version> \
  --go_out=pb --go_opt=paths=source_relative \
  --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
  proto/*.proto
```
