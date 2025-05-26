.PHONY: all build run-server run-gin run-client run-trigger clean help

# Default target
all: build

# Build all binaries
build:
	@echo "Building all binaries..."
	@go build -o bin/server ./cmd/server
	@go build -o bin/gin-server ./cmd/gin-server
	@go build -o bin/subscription-client ./cmd/subscription-client
	@go build -o bin/trigger-events ./cmd/trigger-events
	@echo "All binaries built in ./bin/"

# Run the main GraphQL server
run-server:
	@echo "Starting GraphQL server on port 8080..."
	@go run ./cmd/server

# Run the Gin-based server
run-gin:
	@echo "Starting Gin-based GraphQL server on port 8081..."
	@go run ./cmd/gin-server

# Run the subscription client
run-client:
	@echo "Starting subscription client..."
	@go run ./cmd/subscription-client

# Run the event trigger
run-trigger:
	@echo "Starting event trigger..."
	@go run ./cmd/trigger-events

# Run server and client in parallel (requires GNU parallel or similar)
demo:
	@echo "Starting demo (server + client)..."
	@echo "Start the server first with 'make run-server'"
	@echo "Then in another terminal: 'make run-client'"
	@echo "And in a third terminal: 'make run-trigger'"

# Clean built binaries
clean:
	@echo "Cleaning binaries..."
	@rm -rf bin/

# Show help
help:
	@echo "go-quickgraph-sample Makefile"
	@echo "============================"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build all binaries"
	@echo "  make run-server    - Run the main GraphQL server (port 8080)"
	@echo "  make run-gin       - Run the Gin-based server (port 8081)"
	@echo "  make run-client    - Run the subscription client"
	@echo "  make run-trigger   - Run the event trigger"
	@echo "  make demo          - Instructions for running the full demo"
	@echo "  make clean         - Remove built binaries"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make run-server    (in terminal 1)"
	@echo "  2. make run-client    (in terminal 2)"
	@echo "  3. make run-trigger   (in terminal 3)"