# PowerShell setup script for Mini Sticker Engine

Write-Host "=== Mini Sticker Engine Setup ===" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "Checking Docker..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version 2>&1
    Write-Host "✓ Docker found: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker not found. Please install Docker Desktop." -ForegroundColor Red
    exit 1
}

# Check if Docker daemon is running
Write-Host "Checking Docker daemon..." -ForegroundColor Yellow
try {
    docker ps 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Docker daemon is running" -ForegroundColor Green
    } else {
        Write-Host "✗ Docker daemon is not running. Please start Docker Desktop." -ForegroundColor Red
        Write-Host "  Start Docker Desktop and wait for it to fully start, then run this script again." -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "✗ Docker daemon is not running. Please start Docker Desktop." -ForegroundColor Red
    Write-Host "  Start Docker Desktop and wait for it to fully start, then run this script again." -ForegroundColor Yellow
    exit 1
}

# Create data directory
Write-Host ""
Write-Host "Setting up database directory..." -ForegroundColor Yellow
if (-not (Test-Path "data")) {
    New-Item -ItemType Directory -Path "data" | Out-Null
    Write-Host "✓ Created data directory" -ForegroundColor Green
} else {
    Write-Host "✓ Data directory already exists" -ForegroundColor Green
}

# Build Docker image
Write-Host ""
Write-Host "Building Docker image..." -ForegroundColor Yellow
docker build -t sticker-engine:latest .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Docker image built successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to build Docker image" -ForegroundColor Red
    exit 1
}

# Start with docker-compose
Write-Host ""
Write-Host "Starting application with Docker Compose..." -ForegroundColor Yellow
docker-compose up -d
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Application started successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to start application" -ForegroundColor Red
    exit 1
}

# Wait a moment for the app to start
Write-Host ""
Write-Host "Waiting for application to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Check health
Write-Host ""
Write-Host "Checking application health..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Application is healthy and running!" -ForegroundColor Green
        Write-Host ""
        Write-Host "=== Setup Complete ===" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Application is running at: http://localhost:8080" -ForegroundColor Green
        Write-Host ""
        Write-Host "Useful commands:" -ForegroundColor Yellow
        Write-Host "  View logs:        docker-compose logs -f" -ForegroundColor White
        Write-Host "  Stop application: docker-compose down" -ForegroundColor White
        Write-Host "  Restart:          docker-compose restart" -ForegroundColor White
        Write-Host ""
        Write-Host "Database file: data/stickers.db" -ForegroundColor Cyan
    } else {
        Write-Host "⚠ Application started but health check returned status $($response.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠ Application started but health check failed. It may still be starting up." -ForegroundColor Yellow
    Write-Host "  Try: docker-compose logs" -ForegroundColor White
}

