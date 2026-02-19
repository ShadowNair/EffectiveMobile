APP_NAME=test_task
BIN=bin/app
GO=go

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make test        - run unit tests"
	@echo "  make lint        - run golangci-lint"
	@echo "  make fmt         - gofmt + goimports"
	@echo "  make tidy        - go mod tidy"
	@echo "  make build       - build binary"
	@echo "  make run         - run locally"
	@echo "  make docker-up   - docker compose up --build"
	@echo "  make docker-down - docker compose down -v"

.PHONY: test
test:
	$(GO) test ./... -race -count=1

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -w .
	goimports -w .

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: build
build:
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/app

.PHONY: run
run:
	$(GO) run ./cmd/app

.PHONY: docker-up
docker-up:
	docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down -v
