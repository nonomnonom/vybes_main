# Go parameters
BINARY_NAME=vybes-api
BINARY_UNIX=$(BINARY_NAME)

# Docker parameters
DOCKER_IMAGE_NAME=vybes-api
DOCKER_TAG=latest

# Default target executed when 'make' is run without arguments
.DEFAULT_GOAL := help

.PHONY: all build run test backtest docker-build docker-up docker-down docker-logs help

## build: Compile the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/api

## run: Run the application locally
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./$(BINARY_NAME)

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./...

## backtest: Run a load test on the API
backtest:
	@echo "Running backtest..."
	@bash test/backtest.sh

## docker-build: Build the Docker image for the API
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

## docker-up: Start all services using Docker Compose
docker-up:
	@echo "Starting all services with Docker Compose..."
	@docker-compose up -d --build

## docker-down: Stop all services started with Docker Compose
docker-down:
	@echo "Stopping all services..."
	@docker-compose down

## docker-logs: View logs from all running services
docker-logs:
	@echo "Tailing logs..."
	@docker-compose logs -f

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Compile the application"
	@echo "  run            Run the application locally"
	@echo "  test           Run all tests"
	@echo "  backtest       Run a load test on the API"
	@echo "  docker-build   Build the Docker image for the API"
	@echo "  docker-up      Start all services using Docker Compose"
	@echo "  docker-down    Stop all services started with Docker Compose"
	@echo "  docker-logs    View logs from all running services"