FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates

# Copy necessary files for dependency download
COPY backend/services/auth/go.mod backend/services/auth/go.sum ./
# Copy shared proto files needed by go.mod replace directives BEFORE downloading deps
COPY backend/proto ./backend/proto 

# Download dependencies
RUN go mod download

# Copy the rest of the auth service source code
COPY backend/services/auth/ . 
# Ensure proto is copied again in case source code relies on it directly
# (This might be redundant depending on imports, but safe)
COPY backend/proto ./backend/proto

# Build the binary
# The build command now runs in /app where the service code root is
# Assuming main package is in ./cmd/
RUN go build -o /app/auth-service -ldflags="-w -s" ./cmd

# --- Final Stage ---
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage, changing ownership
# Binary is now built to /app/auth-service in builder stage
COPY --chown=appuser:appgroup --from=builder /app/auth-service .

# Switch to the non-root user
USER appuser

# Expose the port
EXPOSE 9090

# Run the binary
CMD ["./auth-service"]