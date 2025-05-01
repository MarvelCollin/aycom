FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy proto files
COPY proto ./proto/

# Install required packages
RUN go get github.com/lib/pq
RUN go get github.com/dgrijalva/jwt-go
RUN go get github.com/google/uuid
RUN go get golang.org/x/crypto/bcrypt
RUN go get google.golang.org/grpc
RUN go get golang.org/x/net/context

# Build the application
RUN go build -o auth-service .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install required tools
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .
COPY --from=builder /app/proto ./proto

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose the port
EXPOSE 9090

# Use entrypoint script
ENTRYPOINT ["./entrypoint.sh"]