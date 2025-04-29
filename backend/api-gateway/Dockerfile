FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy dependency files and download dependencies first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
# Assuming the main package is in the root of the api-gateway directory
RUN go build -o api-gateway -ldflags="-w -s" .

# --- Final Stage ---
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the built binary from the builder stage, changing ownership
COPY --chown=appuser:appgroup --from=builder /app/api-gateway .

# Switch to the non-root user
USER appuser

# Expose the port the service listens on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./api-gateway"] 