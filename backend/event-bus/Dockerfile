FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git

# Copy Go module files first for better caching
COPY go.mod go.sum ./

# Copy entire codebase
COPY . .

# Clean any existing vendor directory and ensure dependencies are properly downloaded
RUN rm -rf vendor && \
    go mod tidy && \
    go mod download

# Build the service with explicit flag to ignore vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/event-bus .

# Final stage
FROM alpine:latest

# Add certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/event-bus .

# Set environment variables
ENV PORT=8000

# Use non-root user
USER appuser

# Expose ports
EXPOSE 8000

# Start the service
CMD ["./event-bus"] 