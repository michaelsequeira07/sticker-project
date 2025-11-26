#!/bin/bash
# Bash setup script for Mini Sticker Engine

echo "=== Mini Sticker Engine Setup ==="
echo ""

# Check if Docker is installed
echo "Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo "✗ Docker not found. Please install Docker."
    exit 1
fi
echo "✓ Docker found: $(docker --version)"

# Check if Docker daemon is running
echo "Checking Docker daemon..."
if ! docker ps &> /dev/null; then
    echo "✗ Docker daemon is not running. Please start Docker Desktop."
    exit 1
fi
echo "✓ Docker daemon is running"

# Create data directory
echo ""
echo "Setting up database directory..."
mkdir -p data
echo "✓ Data directory ready"

# Build Docker image
echo ""
echo "Building Docker image..."
docker build -t sticker-engine:latest .
if [ $? -ne 0 ]; then
    echo "✗ Failed to build Docker image"
    exit 1
fi
echo "✓ Docker image built successfully"

# Start with docker-compose
echo ""
echo "Starting application with Docker Compose..."
docker-compose up -d
if [ $? -ne 0 ]; then
    echo "✗ Failed to start application"
    exit 1
fi
echo "✓ Application started successfully"

# Wait a moment for the app to start
echo ""
echo "Waiting for application to start..."
sleep 3

# Check health
echo ""
echo "Checking application health..."
if curl -f http://localhost:8080/health &> /dev/null; then
    echo "✓ Application is healthy and running!"
    echo ""
    echo "=== Setup Complete ==="
    echo ""
    echo "Application is running at: http://localhost:8080"
    echo ""
    echo "Useful commands:"
    echo "  View logs:        docker-compose logs -f"
    echo "  Stop application: docker-compose down"
    echo "  Restart:          docker-compose restart"
    echo ""
    echo "Database file: data/stickers.db"
else
    echo "⚠ Application started but health check failed. It may still be starting up."
    echo "  Try: docker-compose logs"
fi

