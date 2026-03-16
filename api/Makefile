BINARY := geokrety-stats-api
CMD := ./cmd/server
GO := go

.PHONY: help fmt lint vet build run test tidy docker-build clean

.PHONY: help fmt lint vet build run test tidy docker-build up down clean

help:
	@echo "Available targets: fmt lint vet build run test tidy docker-build up down clean"

fmt:
	@echo "Running gofmt..."
	@gofmt -w .

lint: fmt
	@echo "Running go vet..."
	@$(GO) vet ./...

vet: lint

build:
	@echo "Building $(BINARY)..."
	@mkdir -p bin
	@$(GO) build -o bin/$(BINARY) $(CMD)
	@echo "Built bin/$(BINARY)"

run:
	@echo "Running server..."
	@$(GO) run $(CMD)

test:
	@echo "Running tests..."
	@$(GO) test ./...

tidy:
	@echo "Tidying modules..."
	@$(GO) mod tidy

docker-build:
	@echo "Building Docker image geokrety/geokrety-stats-api:local"
	docker compose build

up:
	@echo "Starting stack with docker compose (force recreate)..."
	docker compose up --force-recreate -d

down:
	@echo "Stopping stack..."
	docker compose down

clean:
	@echo "Cleaning..."
	@rm -f bin/$(BINARY)
	@echo "Cleaned."
