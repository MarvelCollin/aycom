FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git postgresql-client ca-certificates

# Copy the user service directory and any proto files needed
COPY ./backend/services/user /app/
COPY ./backend/proto /app/backend/proto

# Download dependencies
RUN go mod download

# Build the binary
RUN go build -o user-service .

# Final stage: create a clean image
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates and PostgreSQL client
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN apk add --no-cache postgresql-client

# Copy the binary from the builder stage
COPY --chown=appuser:appgroup --from=builder /app/user-service .

# Switch to the non-root user
USER appuser

# Expose the port
EXPOSE 9091

# Run the binary
ENTRYPOINT ["./user-service"]