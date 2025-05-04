FROM golang:1.24-alpine AS builder

# Build context is ./backend

WORKDIR /app # Base for the backend code

# Copy root go.mod and go.sum first
COPY go.mod go.sum ./
# RUN go mod download # Download root deps

# Copy service-specific go.mod and go.sum
COPY services/thread/go.mod services/thread/go.sum ./services/thread/
# Change to service directory
WORKDIR /app/services/thread

# Copy the rest of the backend source code
WORKDIR /app
COPY . .

# Download dependencies AFTER all code is copied
RUN go mod download
# Sync vendor directory AFTER download
RUN go mod vendor

# Build the service from its directory
WORKDIR /app/services/thread
# Assuming proto is generated locally before build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/thread-service ./main.go

# --- Final Stage ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates # Add CA certs
# Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app # Use /app as final WORKDIR for consistency
# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/thread-service .
# Switch to non-root user
USER appuser
EXPOSE 9092
CMD ["./thread-service"] 