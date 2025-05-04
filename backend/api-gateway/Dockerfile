# Builder stage
FROM golang:1.24-alpine AS builder

# Install only necessary build tools (git for dependencies)
RUN apk add --no-cache git

WORKDIR /app

# Copy root go module files FIRST (relative to ./backend context)
COPY go.mod go.sum /app/
# Download root dependencies
RUN cd /app && go mod download

# Copy the rest of the source code from the context (./backend)
# This will include the pre-generated .pb.go files
COPY . /app

# Tidy and download specific dependencies for the api-gateway module
RUN cd /app/api-gateway && go mod tidy && go mod download

# Build the application from within its directory
RUN cd /app/api-gateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

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
COPY --from=builder --chown=appuser:appgroup /app/api-gateway/app .

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