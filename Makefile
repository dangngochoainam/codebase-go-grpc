start:
	docker compose -f docker-compose.debug.yaml up -d

attach:
	docker exec -it codebase-go-grpc-debug /bin/bash

mod:
	go mod tidy
	go mod vendor

run-watch:
	CompileDaemon -graceful-kill=true -build="go build -v -gcflags all=-N -o ./cmd/server/bin ./cmd/server/main.go" -command="dlv exec ./cmd/server/bin --listen=:2302 --headless=true --api-version=2 --accept-multiclient --continue"

run:
	go run ./cmd/server/main.go

GRPC_GATEWAY_PATH := $(shell find $(shell go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway -maxdepth 1 -name "v2@*" -type d 2>/dev/null | sort -rV | head -1)
gen-proto:
	protoc \
		--proto_path=proto \
		-I $(GRPC_GATEWAY_PATH) \
		-I /usr/local/include \
		--go_out=pb \
		--go_opt=paths=source_relative \
		--go-grpc_out=pb \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb \
		--grpc-gateway_opt=paths=source_relative \
		proto/*.proto
