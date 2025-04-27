FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service .

# Use a smaller image for the final application
FROM alpine:latest

# Install PostgreSQL client and CA certificates
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/user-service .
COPY --from=builder /app/entrypoint.sh .

# Make the entrypoint script executable
RUN chmod +x /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]