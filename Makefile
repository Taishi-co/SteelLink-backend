.PHONY: build run test clean fmt lint deps

# Binary name
BINARY_NAME=steellink-backend
VERSION?=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GO_VERSION=$(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: deps build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the binary
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run with config
run-config:
	go run ./cmd/api -config config/config.yaml

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# View test coverage
coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Build docker image
docker-build:
	docker build -t steellink-backend:$(VERSION) .

# Run in docker
docker-run:
	docker-compose up -d

# Stop docker
docker-stop:
	docker-compose down

# Generate swagger docs
swagger:
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/api/main.go -o docs/swagger

# Development with hot reload
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Database migration
migrate-up:
	go run ./scripts/migrate.go up

migrate-down:
	go run ./scripts/migrate.go down

# Help
help:
	@echo "SteelLink Backend - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deps        Install dependencies"
	@echo "  build       Build the binary"
	@echo "  run         Run the application"
	@echo "  test        Run tests"
	@echo "  coverage    Generate test coverage report"
	@echo "  clean       Clean build artifacts"
	@echo "  fmt         Format code"
	@echo "  lint        Lint code"
	@echo "  docker-build Build docker image"
	@echo "  docker-run  Run with docker-compose"
	@echo "  docker-stop Stop docker-compose"
	@echo "  swagger     Generate swagger docs"
	@echo "  dev         Development with hot reload"
	@echo "  help        Show this help message"
