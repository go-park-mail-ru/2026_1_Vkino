.PHONY: init generate

generate:
	go generate ./internal/app/movie-service/repository/...

PACKAGES_NO_MOCKS := $(shell go list ./... | grep -v '/mocks$$')

init:
	cp .github/hooks/* .git/hooks
	chmod +x .git/hooks/*
	migrate create -ext sql -dir ./migrations migration

MIGRATIONS_DIR := ./migrations

test:
	go test $(PACKAGES_NO_MOCKS) -cover

cover:
	-go test $(PACKAGES_NO_MOCKS) -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm -f coverage.out

cover-total:
	@echo "=== Total project coverage ==="
	@go test $(PACKAGES_NO_MOCKS) -coverprofile=coverage.out > /dev/null 2>&1 || true
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'

run-build:
	docker compose -f deployments/dev/compose.yaml up --build

run-stop:
	docker compose -f deployments/dev/compose.yaml down -v

up:
	docker compose -f deployments/dev/compose.yaml up

down:
	docker compose -f deployments/dev/compose.yaml down

proto-gen:
	protoc -I proto --go_out=./pkg/gen --go_opt=paths=source_relative --go-grpc_out=./pkg/gen --go-grpc_opt=paths=source_relative proto/support/v1/support.proto
	protoc -I proto --go_out=./pkg/gen --go_opt=paths=source_relative --go-grpc_out=./pkg/gen --go-grpc_opt=paths=source_relative proto/movie/v1/movie.proto
	protoc -I proto --go_out=./pkg/gen --go_opt=paths=source_relative --go-grpc_out=./pkg/gen --go-grpc_opt=paths=source_relative proto/user/v1/user.proto
	protoc -I proto --go_out=./pkg/gen --go_opt=paths=source_relative --go-grpc_out=./pkg/gen --go-grpc_opt=paths=source_relative proto/auth/v1/auth.proto
