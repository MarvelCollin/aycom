FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates

# Copy dependency files and proto files
COPY go.mod go.sum ./
COPY ../../proto /app/backend/proto

# Download dependencies
RUN go mod download

# Copy the rest of the auth service code
COPY . .

# Build the binary
RUN go build -o auth-service -ldflags="-w -s" .

# --- Final Stage ---
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage, changing ownership
COPY --chown=appuser:appgroup --from=builder /app/auth-service .

# Switch to the non-root user
USER appuser

# Expose the port
EXPOSE 9090

# Run the binary
CMD ["./auth-service"]