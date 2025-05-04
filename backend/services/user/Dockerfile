# Base image for building the application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git

# Copy Go module files first for better caching
COPY go.mod go.sum ./
COPY services/user/go.mod services/user/go.sum ./services/user/

# Copy entire codebase
COPY . .

# Build from user service directory
WORKDIR /app/services/user

# Clean any existing vendor directory and ensure dependencies are properly downloaded
RUN rm -rf vendor && \
    go mod tidy && \
    go mod download

# Build the service with explicit flag to ignore vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/user-service .

# Final stage
FROM alpine:latest

# Add certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/user-service .

# Set environment variables
ENV PORT=9091

# Use non-root user
USER appuser

# Expose ports
EXPOSE 9091

# Start the service
CMD ["./user-service"] 