PHONY: init

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

run dev:
	docker compose -f deployments/dev/compose.yaml up --build

stop dev:
	docker compose -f deployments/dev/compose.yaml down -v

