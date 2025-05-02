# Builder stage
FROM golang:1.24-alpine AS builder

# Install build dependencies including protobuf compiler
RUN apk add --no-cache git protobuf protobuf-dev

# Set working directory
WORKDIR /app

# First, copy over module files and proto directories
COPY backend/go.mod backend/go.sum ./backend/
COPY backend/services/auth/proto/ ./backend/services/auth/proto/
COPY backend/services/thread/proto/ ./backend/services/thread/proto/
COPY backend/services/user/proto/ ./backend/services/user/proto/
COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./backend/api-gateway/

# Install dependencies
RUN cd backend && go mod download
RUN cd backend/api-gateway && go mod download

# Copy application source
COPY backend/api-gateway/ ./backend/api-gateway/

# Build the application
RUN cd backend/api-gateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM alpine:latest

# Add certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create necessary directories with correct permissions
WORKDIR /app
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/backend/api-gateway/app .

# Copy any additional required files
COPY --from=builder --chown=appuser:appgroup /app/backend/api-gateway/docs ./docs

# Environment variables
ENV GIN_MODE=release

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Run the application
CMD ["./app"]