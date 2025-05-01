FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY ./backend/services/auth/go.mod ./backend/services/auth/go.sum ./

# Copy the proto files first
COPY ./backend/services/auth/proto ./proto/

# Copy the source files
COPY ./backend/services/auth/*.go ./

# Download dependencies
RUN go mod tidy

# Build the application
RUN go build -o auth-service .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Copy binary from builder stage
COPY --from=builder /app/auth-service .

# Expose port
EXPOSE 9090

# Run the service
CMD ["./auth-service"]