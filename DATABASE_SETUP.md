# Database Setup Guide

## Database Overview

The Mini Sticker Engine uses **SQLite** as its database, which is perfect for this exercise because:
- No external database server required
- Auto-creates on first run
- Lightweight and fast
- Easy to backup and inspect

## Database Location

### Local Development
- **File**: `stickers.db` (in project root)
- **Path**: `C:\Users\Carol\sticker_project\stickers.db`

### Docker Deployment
- **File**: `data/stickers.db`
- **Path**: `./data/stickers.db` (mapped to container's `/root/data/stickers.db`)

## Database Schema

The database consists of three tables:

### 1. `transactions`
Stores transaction records with stickers earned.

```sql
CREATE TABLE transactions (
    transaction_id TEXT PRIMARY KEY,
    shopper_id TEXT NOT NULL,
    store_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    total_amount REAL NOT NULL,
    stickers_earned INTEGER NOT NULL,
    created_at TEXT NOT NULL
)
```

### 2. `transaction_items`
Stores individual items within transactions.

```sql
CREATE TABLE transaction_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_id TEXT NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    category TEXT NOT NULL,
    FOREIGN KEY (transaction_id) REFERENCES transactions(transaction_id)
)
```

### 3. `redemptions`
Stores redemption history.

```sql
CREATE TABLE redemptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shopper_id TEXT NOT NULL,
    reward_name TEXT NOT NULL,
    stickers_cost INTEGER NOT NULL,
    redeemed_at TEXT NOT NULL
)
```

## Initialization

The database is **automatically created** when you start the application:

```bash
go run main.go
```

The `initDB()` function in `main.go` creates all tables if they don't exist.

## Configuration

### Environment Variable

You can customize the database path using the `DB_PATH` environment variable:

**Windows PowerShell:**
```powershell
$env:DB_PATH="C:\path\to\custom\stickers.db"
go run main.go
```

**Linux/Mac:**
```bash
export DB_PATH=/path/to/custom/stickers.db
go run main.go
```

**Docker:**
```bash
docker run -e DB_PATH=/custom/path/stickers.db ...
```

## Database Operations

### Verify Database Exists

```powershell
# Windows PowerShell
Test-Path "stickers.db"
Get-Item "stickers.db"

# Linux/Mac
ls -lh stickers.db
```

### Inspect Database

You can use SQLite command-line tools or GUI tools:

**SQLite CLI:**
```bash
sqlite3 stickers.db

# Then run SQL commands:
.tables
.schema transactions
SELECT * FROM transactions LIMIT 5;
```

**GUI Tools:**
- DB Browser for SQLite (https://sqlitebrowser.org/)
- DBeaver (https://dbeaver.io/)
- VS Code SQLite extension

### Backup Database

```powershell
# Windows PowerShell
Copy-Item "stickers.db" "stickers.db.backup"

# Linux/Mac
cp stickers.db stickers.db.backup
```

### Reset Database

**⚠️ Warning: This will delete all data!**

```powershell
# Stop the application first, then:
Remove-Item "stickers.db"
# Restart the application to recreate it
```

## Testing Database

### Quick Test

1. Start the application:
   ```bash
   go run main.go
   ```

2. Create a test transaction:
   ```powershell
   $body = '{"transaction_id":"tx-test-1","shopper_id":"shopper-test","store_id":"store-01","timestamp":"2025-01-10T10:15:00Z","items":[{"sku":"SKU-1","name":"Test","quantity":1,"unit_price":10,"category":"grocery"}]}'
   Invoke-RestMethod -Uri http://localhost:8080/transactions -Method POST -Body $body -ContentType "application/json"
   ```

3. Verify data:
   ```powershell
   Invoke-RestMethod -Uri http://localhost:8080/shoppers/shopper-test
   ```

### Run Database Test Script

```powershell
.\verify_db.ps1
```

## Docker Database Setup

When using Docker, the database is persisted in a volume:

```yaml
volumes:
  - ./data:/root/data
```

This means:
- Database file: `./data/stickers.db` on your host machine
- Database persists even when container stops
- You can backup by copying the `data` directory

## Production Considerations

For production deployments, consider:

1. **PostgreSQL/MySQL**: Replace SQLite with a production database
2. **Backups**: Implement automated backup strategy
3. **Migrations**: Use migration tools for schema changes
4. **Connection Pooling**: Configure appropriate pool sizes
5. **Monitoring**: Monitor database size and performance

## Troubleshooting

### Database Locked Error
- **Cause**: Multiple processes accessing the database
- **Solution**: Ensure only one instance of the application is running

### Permission Denied
- **Cause**: Insufficient file permissions
- **Solution**: 
  ```powershell
  # Windows
  icacls stickers.db /grant Users:F
  
  # Linux/Mac
  chmod 644 stickers.db
  ```

### Database Not Found
- **Cause**: Application can't create database file
- **Solution**: Check write permissions in the directory

### Corrupted Database
- **Cause**: Unexpected shutdown or disk issues
- **Solution**: Restore from backup or recreate database

## Current Status

✅ Database file created: `stickers.db`
✅ Schema initialized with all tables
✅ Database operations working correctly
✅ Ready for use

