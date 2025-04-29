FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git postgresql-client ca-certificates

# Copy necessary files for dependency download
COPY backend/services/user/go.mod backend/services/user/go.sum ./
# Copy shared proto files needed by go.mod replace directives BEFORE downloading deps
COPY backend/proto ./backend/proto

# Download dependencies
RUN go mod download

# Copy the rest of the user service source code
COPY backend/services/user/ .
# Ensure proto is copied again in case source code relies on it directly
# (This might be redundant depending on imports, but safe)
COPY backend/proto ./backend/proto

# Build the binary
# The build command now runs in /app where the service code root is
# Assuming main package is in ./cmd/
RUN go build -o /app/user-service ./cmd

# Final stage: create a clean image
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates and PostgreSQL client
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN apk add --no-cache postgresql-client

# Copy the binary from the builder stage
# Binary is now built to /app/user-service in builder stage
COPY --chown=appuser:appgroup --from=builder /app/user-service .

# Switch to the non-root user
USER appuser

# Expose the port
EXPOSE 9091

# Run the binary
ENTRYPOINT ["./user-service"]