.PHONY: help build run test test-integration fmt vet lint docker-build docker-run clean

BINARY_NAME=potato-nice-thelma
DOCKER_IMAGE=potato-nice-thelma

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o bin/$(BINARY_NAME) ./cmd/server

run: build ## Build and run the server
	./bin/$(BINARY_NAME)

test: ## Run unit tests with race detector
	go test -race ./...

test-integration: ## Run integration tests with race detector
	go test -race -tags=integration ./...

fmt: ## Format all Go source files
	go fmt ./...

vet: ## Run go vet on all packages
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

docker-build: ## Build the Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run the Docker container
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

clean: ## Remove build artifacts
	rm -rf bin/
