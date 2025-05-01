FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy source code
COPY . .

# Install github.com/lib/pq explicitly
RUN go get github.com/lib/pq

# Build the application
RUN go build -o auth-service .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install required tools
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Expose the port
EXPOSE 9090

# Run the service directly
CMD ["./auth-service"]