# Deployment Guide

This guide covers different deployment options for the Mini Sticker Engine.

## Local Development

### Prerequisites
- Go 1.22 or higher
- Git (optional)

### Quick Start

```bash
# Install dependencies
go mod download

# Run the application
go run main.go
```

The server will start on `http://localhost:8080`.

## Docker Deployment

### Option 1: Docker Compose (Recommended)

Docker Compose handles building, running, and volume management automatically.

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

The database will be persisted in `./data/stickers.db`.

### Option 2: Docker CLI

```bash
# Build image
docker build -t sticker-engine:latest .

# Run container
docker run -d \
  --name sticker-engine \
  -p 8080:8080 \
  -v $(pwd)/data:/root/data \
  -e PORT=8080 \
  -e DB_PATH=/root/data/stickers.db \
  sticker-engine:latest
```

### Option 3: Using Makefile

```bash
# Build Docker image
make docker-build

# Run with docker-compose
make docker-run

# View logs
make docker-logs

# Stop
make docker-stop
```

## Cloud Deployment

### Heroku

1. Create a `Procfile`:
```
web: ./sticker-engine
```

2. Deploy:
```bash
heroku create your-app-name
git push heroku main
```

**Note:** Heroku uses ephemeral filesystem, so you'll need to use a persistent database like PostgreSQL instead of SQLite.

### AWS (EC2/ECS)

1. Build and push Docker image to ECR
2. Create ECS task definition
3. Configure RDS or use persistent EBS volume for SQLite

### Google Cloud Platform

1. Build container image:
```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/sticker-engine
```

2. Deploy to Cloud Run:
```bash
gcloud run deploy sticker-engine \
  --image gcr.io/PROJECT_ID/sticker-engine \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

**Note:** Cloud Run is serverless and uses ephemeral storage. Consider using Cloud SQL or Cloud Storage for database persistence.

### Production Considerations

For production deployments, consider:

1. **Database**: Replace SQLite with PostgreSQL, MySQL, or another production database
2. **Environment Variables**: Use secrets management (AWS Secrets Manager, HashiCorp Vault, etc.)
3. **Logging**: Configure structured logging to a log aggregation service (CloudWatch, Datadog, etc.)
4. **Monitoring**: Add health checks and metrics endpoints
5. **Scaling**: Use load balancers and multiple instances
6. **Backups**: Implement database backup strategy
7. **Security**: 
   - Enable HTTPS/TLS
   - Add authentication/authorization
   - Implement rate limiting
   - Use parameterized queries (already implemented)

## Database Migration

Currently, the database schema is created automatically on first run. For production:

1. Create migration scripts
2. Use a migration tool (golang-migrate, etc.)
3. Run migrations as part of deployment process

Example migration structure:
```
migrations/
  ├── 001_create_transactions.up.sql
  ├── 001_create_transactions.down.sql
  ├── 002_create_transaction_items.up.sql
  └── 002_create_transaction_items.down.sql
```

## Health Checks

The application includes a health check endpoint:

```bash
curl http://localhost:8080/health
```

Use this for:
- Docker health checks
- Load balancer health checks
- Kubernetes liveness/readiness probes

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `stickers.db` | Database file path |
| `GIN_MODE` | `debug` | Gin framework mode (`debug` or `release`) |

## Troubleshooting

### Database Locked Error

If you see "database is locked" errors:
- Ensure only one instance is accessing the database
- For production, use a proper database server (PostgreSQL, etc.)

### Port Already in Use

Change the port:
```bash
PORT=3000 go run main.go
```

Or in Docker:
```bash
docker run -p 3000:3000 -e PORT=3000 sticker-engine:latest
```

### Permission Denied

Ensure the application has write permissions for the database directory:
```bash
chmod 755 data/
```

