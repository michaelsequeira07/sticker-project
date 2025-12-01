# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY cache/ ./cache/
COPY auth/ ./auth/
COPY handlers/ ./handlers/
COPY database/ ./database/
COPY models/ ./models/
COPY calculator/ ./calculator/
COPY validation/ ./validation/
COPY config/ ./config/

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o sticker-engine main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/sticker-engine .

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/root/data/stickers.db
ENV GIN_MODE=release

# Run the application
CMD ["./sticker-engine"]

