FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git

# Copy Go module files first for better caching
COPY go.mod go.sum ./
COPY services/community/go.mod services/community/go.sum ./services/community/

# Copy entire codebase
COPY . .

# Build from community service directory
WORKDIR /app/services/community

# Clean any existing vendor directory and ensure dependencies are properly downloaded
RUN rm -rf vendor && \
    go mod tidy && \
    go mod download

# Build the service with explicit flag to ignore vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/community-service .

# Final stage
FROM alpine:latest

# Add certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/community-service .

# Set environment variables
ENV PORT=9093

# Use non-root user
USER appuser

# Expose ports
EXPOSE 9093

# Start the service
CMD ["./community-service"] 