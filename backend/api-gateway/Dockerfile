# Builder stage
FROM golang:1.24-alpine AS builder

# Install only necessary build tools (git for dependencies)
RUN apk add --no-cache git

WORKDIR /go/src/aycom/backend

# Copy root go module files FIRST (relative to ./backend context)
COPY go.mod go.sum ./
# Download root dependencies
# RUN go mod download

# Copy the rest of the source code from the context (./backend)
# This will include the pre-generated .pb.go files
COPY . .

# Download dependencies AFTER all code is copied
RUN go mod download
# Sync vendor directory AFTER download
RUN go mod vendor

# Build the application from the root, targeting the api-gateway package
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api-gateway-app ./api-gateway

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
COPY --from=builder --chown=appuser:appgroup /app/api-gateway-app /app/app

# (Optional) Copy Swagger docs if present
# COPY --from=builder --chown=appuser:appgroup /app/backend/api-gateway/docs ./docs

# Environment variables
ENV GIN_MODE=debug

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Run the application
CMD ["./app"]