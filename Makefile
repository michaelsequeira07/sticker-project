.PHONY: build run test clean docker-build docker-run docker-stop help

# Build the application
build:
	go build -o sticker-engine main.go

# Run the application locally
run:
	go run main.go

# Run tests
test:
	go test -v

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f sticker-engine sticker-engine.exe
	rm -f *.db *.sqlite *.sqlite3
	rm -f coverage.out coverage.html
	rm -rf data/

# Build Docker image
docker-build:
	docker build -t sticker-engine:latest .

# Run Docker container
docker-run:
	docker-compose up -d

# Stop Docker container
docker-stop:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f

# Run Docker container interactively
docker-run-interactive:
	docker run -it --rm -p 8080:8080 -v $(PWD)/data:/root/data sticker-engine:latest

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application binary"
	@echo "  run                - Run the application locally"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  clean              - Remove build artifacts and databases"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container (docker-compose)"
	@echo "  docker-stop        - Stop Docker container"
	@echo "  docker-logs        - View Docker logs"
	@echo "  docker-run-interactive - Run Docker container interactively"
	@echo "  fmt                - Format code"
	@echo "  deps               - Download and tidy dependencies"
	@echo "  help               - Show this help message"

