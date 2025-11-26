# Mini Sticker Engine

A simplified sticker-style loyalty campaign system where shoppers earn "stickers" when they make purchases and can redeem them for rewards.

Built with **Go** and **Gin** web framework.

## Features

### MVP Requirements
- ✅ **Transaction Ingestion**: Submit transactions via HTTP API
- ✅ **Sticker Calculation**: Applies campaign rules to calculate stickers earned
- ✅ **Shopper Status**: View shopper balance and transaction history
- ✅ **Idempotency**: Handles duplicate transaction IDs gracefully
- ✅ **Input Validation**: Validates transaction data and provides clear error messages

### Stretch Goals Implemented
- ✅ **Santa's Helper**: Sticker redemption for rewards (Mug: 10 stickers, Tote bag: 20 stickers)
- ✅ **The Analyst**: Statistics endpoint showing total stickers, transactions, and per-store breakdown
- ✅ **The Detective**: Debug endpoint and structured logging for tracing transactions with full calculation breakdown
- ✅ **The Tester**: Comprehensive unit tests for calculation logic, validation, and API endpoints

## Campaign Rules

1. **Base earn rate**: 1 sticker per $10 of total basket spend
2. **Promo item bonus**: +1 extra sticker per unit for items with `category = "promo"`
3. **Per-transaction cap**: Maximum 5 stickers per transaction

## Setup

### Prerequisites
- Go 1.22 or higher
- Git (optional, for cloning)

### Installation

1. Clone or download this repository

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

Or use the Makefile:
```bash
make run
```

The server will start on `http://0.0.0.0:8080` (or the port specified by the `PORT` environment variable).

The database (`stickers.db`) will be automatically created on first run.

### Quick Start with Docker

```bash
# Using Docker Compose (recommended)
docker-compose up -d

# Or using Makefile
make docker-build
make docker-run
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## API Endpoints

### Health Check
```
GET /health
```
Returns server status.

### Create Transaction
```
POST /transactions
Content-Type: application/json

{
  "transaction_id": "tx-1001",
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
    },
    {
      "sku": "SKU-PLUSH",
      "name": "Promo Plush Toy",
      "quantity": 1,
      "unit_price": 15,
      "category": "promo"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "transaction_id": "tx-1001",
  "shopper_id": "shopper-123",
  "stickers_earned": 2,
  "current_balance": 2
}
```

### Get Shopper Status
```
GET /shoppers/<shopper_id>
```

**Response (200 OK):**
```json
{
  "shopper_id": "shopper-123",
  "current_balance": 5,
  "total_earned": 10,
  "total_redeemed": 5,
  "transactions": [
    {
      "transaction_id": "tx-1001",
      "store_id": "store-01",
      "timestamp": "2025-01-10T10:15:00Z",
      "total_amount": 25.0,
      "stickers_earned": 2
    }
  ],
  "redemptions": [
    {
      "reward_name": "Mug",
      "stickers_cost": 10,
      "redeemed_at": "2025-01-10T11:00:00Z"
    }
  ]
}
```

### Redeem Stickers
```
POST /redemptions
Content-Type: application/json

{
  "shopper_id": "shopper-123",
  "reward_name": "Mug"
}
```

**Available rewards:**
- `Mug`: 10 stickers
- `Tote bag`: 20 stickers

**Response (201 Created):**
```json
{
  "id": 1,
  "shopper_id": "shopper-123",
  "reward_name": "Mug",
  "stickers_cost": 10,
  "previous_balance": 15,
  "new_balance": 5,
  "redeemed_at": "2025-01-10T11:00:00Z"
}
```

### Get Statistics
```
GET /stats
```

**Response (200 OK):**
```json
{
  "total_stickers_awarded": 150,
  "total_transactions": 25,
  "total_redemptions": 5,
  "total_stickers_redeemed": 50,
  "stickers_per_store": [
    {
      "store_id": "store-01",
      "total_stickers": 75,
      "transaction_count": 12
    }
  ]
}
```

### Get Transaction Details
```
GET /transactions/<transaction_id>
```

**Response (200 OK):**
```json
{
  "transaction_id": "tx-1001",
  "shopper_id": "shopper-123",
  "store_id": "store-01",
  "timestamp": "2025-01-10T10:15:00Z",
  "total_amount": 25.0,
  "stickers_earned": 3,
  "created_at": "2025-01-10T10:15:01Z",
  "items": [
    {
      "sku": "SKU-MILK",
      "name": "Milk",
      "quantity": 2,
      "unit_price": 5,
      "category": "grocery"
    },
    {
      "sku": "SKU-PLUSH",
      "name": "Promo Plush Toy",
      "quantity": 1,
      "unit_price": 15,
      "category": "promo"
    }
  ],
  "calculation": {
    "total_amount": 25.0,
    "base_stickers": 2,
    "promo_bonus": 1,
    "before_cap": 3,
    "after_cap": 3,
    "capped": false
  }
}
```

### Debug Transaction (The Detective)
```
GET /debug/transactions/<transaction_id>
```

This endpoint provides comprehensive debug information about a transaction, including:
- Full transaction details and items
- Complete calculation breakdown
- Shopper balance after this transaction
- Processing timestamp

**Response (200 OK):**
```json
{
  "transaction_id": "tx-1001",
  "shopper_id": "shopper-123",
  "store_id": "store-01",
  "timestamp": "2025-01-10T10:15:00Z",
  "total_amount": 25.0,
  "stickers_earned": 3,
  "created_at": "2025-01-10T10:15:01Z",
  "items": [...],
  "calculation": {
    "total_amount": 25.0,
    "base_stickers": 2,
    "promo_bonus": 1,
    "before_cap": 3,
    "after_cap": 3,
    "capped": false
  },
  "shopper_balance_after": 3,
  "processing_time": "2025-01-10T10:15:01Z"
}
```

**Note:** All transaction processing is logged to stdout with structured logging. You can trace transactions by searching logs for `[TRANSACTION]` entries.

## Running Tests

Run the unit tests with:
```bash
go test -v
```

The tests cover:
- Sticker calculation logic (base rate, promo bonus, cap)
- Input validation
- API endpoints (transaction creation, idempotency, shopper status, redemption)
- Edge cases

## Building

Build a binary:
```bash
go build -o sticker-engine main.go
```

Run the binary:
```bash
./sticker-engine
```

## Database

The application uses SQLite with the following schema:

- **transactions**: Stores transaction records with stickers earned
- **transaction_items**: Stores individual items within transactions
- **redemptions**: Stores redemption history

The database file `stickers.db` is created automatically on first run.

## Error Handling

The API returns appropriate HTTP status codes:
- `200`: Success
- `201`: Created
- `400`: Bad Request (validation errors, insufficient stickers, etc.)
- `404`: Not Found
- `500`: Internal Server Error

Error responses include a JSON object with an `error` field describing the issue.

## Docker Deployment

### Using Docker Compose (Recommended)

1. Build and run with Docker Compose:
```bash
docker-compose up -d
```

2. View logs:
```bash
docker-compose logs -f
```

3. Stop the service:
```bash
docker-compose down
```

The database will be persisted in the `./data` directory.

### Using Docker directly

1. Build the image:
```bash
docker build -t sticker-engine:latest .
```

2. Run the container:
```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/root/data \
  -e PORT=8080 \
  sticker-engine:latest
```

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DB_PATH`: Database file path (default: stickers.db)
- `GIN_MODE`: Gin mode - `debug` or `release` (default: debug)

## Project Structure

```
.
├── main.go              # Main application and handlers
├── main_test.go         # Unit tests
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Docker build configuration
├── docker-compose.yml   # Docker Compose configuration
├── Makefile             # Common build/run commands
├── .dockerignore        # Docker ignore patterns
├── README.md            # This file
└── TECH_NOTES.md        # Technical notes and examples
```
