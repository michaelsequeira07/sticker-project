# Technical Notes

## Example Commands / Requests

### Using curl

#### 1. Create a transaction
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

#### 2. Get shopper status
```bash
curl http://localhost:8080/shoppers/shopper-123
```

#### 3. Redeem stickers
```bash
curl -X POST http://localhost:8080/redemptions \
  -H "Content-Type: application/json" \
  -d '{
    "shopper_id": "shopper-123",
    "reward_name": "Mug"
  }'
```

#### 4. Get statistics
```bash
curl http://localhost:8080/stats
```

#### 5. Get transaction details
```bash
curl http://localhost:8080/transactions/tx-1001
```

#### 6. Debug transaction (The Detective)
```bash
curl http://localhost:8080/debug/transactions/tx-1001
```

This endpoint shows:
- Full transaction details
- Complete calculation breakdown (base stickers, promo bonus, cap application)
- Shopper balance after this transaction
- Processing timestamp

### Using Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Create transaction
    tx := map[string]interface{}{
        "transaction_id": "tx-1001",
        "shopper_id":     "shopper-123",
        "store_id":       "store-01",
        "timestamp":      "2025-01-10T10:15:00Z",
        "items": []map[string]interface{}{
            {
                "sku":      "SKU-MILK",
                "name":     "Milk",
                "quantity": 2,
                "unit_price": 5,
                "category": "grocery",
            },
            {
                "sku":      "SKU-PLUSH",
                "name":     "Promo Plush Toy",
                "quantity": 1,
                "unit_price": 15,
                "category": "promo",
            },
        },
    }
    
    jsonData, _ := json.Marshal(tx)
    resp, _ := http.Post("http://localhost:8080/transactions", 
        "application/json", bytes.NewBuffer(jsonData))
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result)
}
```

## Logging and Tracing (The Detective)

The application includes structured logging to help trace transaction processing:

### Log Format

All transaction-related logs are prefixed with `[TRANSACTION]` and include:
- Transaction ID
- Shopper ID
- Store ID
- Total amount
- Stickers earned
- Calculation breakdown (base stickers, promo bonus, cap status)

### Example Log Output

When processing a transaction, you'll see logs like:
```
[TRANSACTION] Processing transaction_id=tx-1001 shopper_id=shopper-123 store_id=store-01 total_amount=25.00 stickers_earned=3 (base=2 promo=1 before_cap=3 capped=false)
[TRANSACTION] Successfully processed transaction_id=tx-1001 shopper_id=shopper-123 stickers_earned=3
```

For duplicate transactions:
```
[TRANSACTION] Duplicate transaction_id=tx-1001 detected, returning existing result (stickers_earned=3)
```

### Tracing a Transaction

To trace what happened for a specific `transaction_id`:

1. **Search logs**: `grep "tx-1001" logs.txt` (or use your log aggregation tool)
2. **Use debug endpoint**: `GET /debug/transactions/tx-1001` shows full calculation breakdown
3. **Use transaction details**: `GET /transactions/tx-1001` shows transaction with calculation

The debug endpoint provides the most comprehensive view, including:
- All items in the transaction
- Step-by-step calculation (base stickers, promo bonus, cap application)
- Shopper balance after this transaction
- Processing timestamp

## AI Tools Usage

I used AI assistance (Cursor) to help with:
- Initial project structure and boilerplate code
- Code review and suggestions for best practices
- Debugging assistance during development

However, I understand all the code I've written, including:
- The sticker calculation algorithm and business logic
- Database schema design and SQL queries
- API endpoint structure and error handling
- Test cases and validation logic

I'm prepared to explain any part of the codebase, modify it, or debug issues during the review session.

## Design Decisions

### Technology Choices
- **Go**: Fast, efficient, great for backend services, strong typing
- **Gin**: Lightweight web framework, high performance, easy to use
- **SQLite**: No external dependencies, perfect for a self-contained exercise
- **modernc.org/sqlite**: Pure Go SQLite driver, no CGO dependencies

### Architecture
- **Single main.go file**: For simplicity, but with clear separation of concerns:
  - Database initialization
  - Business logic (sticker calculation)
  - Validation
  - API handlers
- **Database schema**: Normalized with separate tables for transactions, items, and redemptions
- **Struct-based models**: Type-safe data structures for requests and responses

### Key Features
1. **Idempotency**: Check for existing transaction_id before inserting to prevent double-awarding
2. **Balance calculation**: Computed on-the-fly from transactions minus redemptions
3. **Input validation**: Comprehensive validation with clear error messages
4. **Error handling**: Proper HTTP status codes and JSON error responses
5. **Database transactions**: Use SQL transactions to ensure atomicity

### Trade-offs
- **In-memory vs Database**: Chose SQLite for persistence and to demonstrate database skills, even though in-memory would be simpler
- **Single file vs modules**: Kept everything in one file for simplicity, but in production would split into packages
- **Synchronous processing**: No background jobs, but transactions are processed immediately
- **No ORM**: Used raw SQL for simplicity and control, but in production might use an ORM like GORM

## Testing Strategy

Unit tests cover:
- Core business logic (sticker calculation with all rules)
- Input validation edge cases
- API endpoint behavior
- Idempotency handling
- Redemption logic

Tests use a separate test database that's cleaned up after each test run. The `testify` library is used for assertions.

## Future Improvements (if this were production)

1. **Performance**:
   - Add database indexes on frequently queried fields (shopper_id, transaction_id)
   - Consider caching for shopper balances using Redis
   - Connection pooling (already handled by database/sql)
   - Background job processing for high-volume scenarios

2. **Scalability**:
   - Move to PostgreSQL or another production database
   - Add read replicas for shopper status queries
   - Horizontal scaling with load balancer
   - Message queue for async processing

3. **Features**:
   - Admin API for managing rewards
   - Webhook notifications for redemptions
   - Campaign rule configuration (instead of hard-coded)
   - Time-based rules (e.g., weekday bonuses)
   - Rate limiting per shopper

4. **Reliability**:
   - Add structured logging (e.g., zap or logrus)
   - Metrics and monitoring (Prometheus)
   - Health checks with database connectivity
   - Graceful shutdown
   - Database migrations system

5. **Security**:
   - Authentication and authorization (JWT)
   - Rate limiting
   - Input sanitization
   - SQL injection prevention (already using parameterized queries)
   - HTTPS/TLS

6. **Code Organization**:
   - Split into packages (models, handlers, services, database)
   - Use dependency injection
   - Configuration management (viper)
   - Separate business logic from HTTP handlers

7. **Go-specific**:
   - Context propagation for request cancellation
   - Structured errors with error wrapping
   - Interface-based design for testability
   - Go modules for dependency management (already using)

## Go Version

This project requires Go 1.22 or higher. The code uses:
- Generics (not used, but compatible)
- Improved error handling
- Modern standard library features

## Dependencies

- `github.com/gin-gonic/gin`: Web framework
- `github.com/stretchr/testify`: Testing assertions
- `modernc.org/sqlite`: Pure Go SQLite driver

All dependencies are managed through Go modules (`go.mod`).
