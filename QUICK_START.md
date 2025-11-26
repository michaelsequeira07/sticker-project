# Quick Start Guide

## Prerequisites Check

✅ Docker Desktop is installed (version 28.3.0)
✅ Docker Compose is installed (version 2.38.1)

## Step 1: Start Docker Desktop

**Important:** Docker Desktop must be running before proceeding.

1. Open Docker Desktop application
2. Wait for it to fully start (you'll see "Docker Desktop is running" in the system tray)
3. Verify it's running by checking the Docker icon in your system tray

## Step 2: Run Setup Script

### Windows (PowerShell):
```powershell
.\setup.ps1
```

### Linux/Mac (Bash):
```bash
chmod +x setup.sh
./setup.sh
```

### Or Manual Setup:

1. **Create data directory** (for database persistence):
   ```powershell
   # Windows PowerShell
   if (-not (Test-Path "data")) { New-Item -ItemType Directory -Path "data" }
   
   # Linux/Mac
   mkdir -p data
   ```

2. **Build Docker image**:
   ```bash
   docker build -t sticker-engine:latest .
   ```

3. **Start the application**:
   ```bash
   docker-compose up -d
   ```

4. **Verify it's running**:
   ```bash
   curl http://localhost:8080/health
   # Or in PowerShell:
   Invoke-WebRequest -Uri http://localhost:8080/health
   ```

## Step 3: Verify Setup

### Check if application is running:
```bash
docker-compose ps
```

### View logs:
```bash
docker-compose logs -f
```

### Test the API:
```bash
# Health check
curl http://localhost:8080/health

# Create a test transaction
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "tx-test-1",
    "shopper_id": "shopper-123",
    "store_id": "store-01",
    "timestamp": "2025-01-10T10:15:00Z",
    "items": [
      {
        "sku": "SKU-MILK",
        "name": "Milk",
        "quantity": 2,
        "unit_price": 5,
        "category": "grocery"
      }
    ]
  }'
```

## Database Location

The database file will be created at:
- **Docker**: `./data/stickers.db` (persisted in the `data` directory)
- **Local**: `stickers.db` (in current directory)

## Common Commands

```bash
# Start application
docker-compose up -d

# Stop application
docker-compose down

# View logs
docker-compose logs -f

# Restart application
docker-compose restart

# Rebuild and restart
docker-compose up -d --build

# Check status
docker-compose ps
```

## Troubleshooting

### Docker Desktop not running
- **Error**: `error during connect: The system cannot find the file specified`
- **Solution**: Start Docker Desktop and wait for it to fully initialize

### Port already in use
- **Error**: `Bind for 0.0.0.0:8080 failed: port is already allocated`
- **Solution**: Change the port in `docker-compose.yml` or stop the service using port 8080

### Database permission issues
- **Error**: `database is locked` or permission denied
- **Solution**: Ensure the `data` directory has proper permissions:
  ```powershell
  # Windows
  icacls data /grant Users:F
  
  # Linux/Mac
  chmod 755 data
  ```

### View detailed logs
```bash
docker-compose logs sticker-engine
```

## Next Steps

Once the application is running:
1. Test the API endpoints (see README.md for full API documentation)
2. Check the database file at `data/stickers.db`
3. View logs to see transaction processing

## Stopping the Application

```bash
docker-compose down
```

This will stop the container but **preserve the database** in the `data` directory.

