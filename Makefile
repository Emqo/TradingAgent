.PHONY: build run test clean docker-build docker-up docker-down lint help

# Variables
BINARY=bin/agent
DOCKER_IMAGE=trading-agent
DOCKER_TAG=latest

# Build the agent binary
build:
	go build -o $(BINARY) ./cmd/agent

# Run the agent
run: build
	./$(BINARY)

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f agent

# Lint the code
lint:
	golangci-lint run ./...

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the agent binary"
	@echo "  run          - Build and run the agent"
	@echo "  test         - Run all tests"
	@echo "  test-cover   - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start Docker services"
	@echo "  docker-down  - Stop Docker services"
	@echo "  docker-logs  - View Docker logs"
	@echo "  lint         - Lint the code"
	@echo "  fmt          - Format the code"
	@echo "  vet          - Vet the code"
	@echo "  tidy         - Tidy dependencies"
	@echo "  help         - Show this help"
