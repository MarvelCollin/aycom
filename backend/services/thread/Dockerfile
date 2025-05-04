FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git

# Copy Go module files first for better caching
COPY go.mod go.sum ./
COPY services/thread/go.mod services/thread/go.sum ./services/thread/

# Set up the directory structure for Go modules
RUN mkdir -p services/user/proto

# Copy entire codebase
COPY . .

# Build from thread service directory
WORKDIR /app/services/thread

# Clean any existing vendor directory and ensure dependencies are properly downloaded
RUN rm -rf vendor && \
    go mod tidy && \
    go mod vendor

# Build the service using vendor directory
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/thread-service .

# Final stage
FROM alpine:latest

# Add certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/thread-service .

# Set environment variables
ENV PORT=9092
ENV HEALTH_PORT=8082

# Use non-root user
USER appuser

# Expose ports
EXPOSE 9092 8082

# Start the service
CMD ["./thread-service"] 