FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy necessary files for dependency download
COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./
# Copy shared proto files needed by go.mod replace directives BEFORE downloading deps
COPY backend/proto ./backend/proto 

# Download dependencies
RUN go mod download

# Copy the rest of the api-gateway source code
COPY backend/api-gateway/ .
# Ensure proto is copied again in case source code relies on it directly
# (This might be redundant depending on imports, but safe)
COPY backend/proto ./backend/proto

# Build the Go app
# The build command now runs in /app where the service code root is
# Assuming main package is in ./cmd/
RUN go build -o /app/api-gateway -ldflags="-w -s" ./cmd

# --- Final Stage ---
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the built binary from the builder stage, changing ownership
# Binary is now built to /app/api-gateway in builder stage
COPY --chown=appuser:appgroup --from=builder /app/api-gateway .

# Switch to the non-root user
USER appuser

# Expose the port the service listens on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./api-gateway"] 