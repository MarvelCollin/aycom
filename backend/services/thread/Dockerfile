# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required dependencies
RUN apk add --no-cache gcc musl-dev git

# Create directories for the service code
RUN mkdir -p /app/proto

# Copy just the proto files first 
COPY backend/services/thread/proto/ /app/proto/

# Copy the rest of the code
COPY backend/services/thread/ .

# Create a go.mod file that doesn't require external GitHub access
RUN rm go.mod go.sum && \
    go mod init thread-service && \
    go get google.golang.org/grpc@v1.59.0 && \
    go get google.golang.org/protobuf@v1.31.0 && \
    go get gorm.io/driver/postgres@v1.5.4 && \
    go get gorm.io/gorm@v1.25.5 && \
    go get github.com/google/uuid@v1.4.0 && \
    go mod tidy

# Build the service
RUN go build -o thread-service .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/thread-service .

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Make sure the application user owns the files
RUN chown -R appuser:appgroup /app

# Use non-root user
USER appuser

# Expose gRPC port
EXPOSE 9092

# Default environment variables (will be overridden by docker-compose or kubernetes)
ENV THREAD_SERVICE_PORT=9092
# DO NOT set sensitive defaults here
# These will be set in docker-compose or kubernetes config
# ENV DB_HOST=
# ENV DB_PORT=
# ENV DB_USER=
# ENV DB_PASSWORD=
# ENV DB_NAME=
# ENV DB_SSLMODE=

# Health check to ensure service is responsive
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD wget -qO- http://localhost:9092/health || exit 1

# Run the service
ENTRYPOINT ["./thread-service"] 