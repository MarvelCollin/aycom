# Base image for building the application
FROM golang:1.24-alpine AS builder

# Set environment variables for Go tools
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Install necessary tools: git, protobuf, and Go protobuf/gRPC plugins
RUN apk add --no-cache git protobuf protobuf-dev && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app # Base for the backend code

# Copy root go module files and download dependencies first
COPY go.mod go.sum /app/

# Copy service-specific go.mod and go.sum
COPY services/user/go.mod services/user/go.sum ./services/user/
# Change to service directory
WORKDIR /app/services/user

# Copy the rest of the application code from the build context (./backend)
WORKDIR /app
COPY . .

# Download dependencies AFTER all code is copied
RUN go mod download
# Sync vendor directory AFTER download
RUN go mod vendor

# Generate proto code from service directory
WORKDIR /app/services/user
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/user.proto

# Build the application from service directory
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./main.go

# --- Final Stage ---
# Use a minimal base image for the final container
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder --chown=appuser:appgroup /app/user-service .

# Expose the port the service runs on
EXPOSE 9091

# Switch to non-root user
USER appuser

# Command to run the application
ENTRYPOINT ["/app/user-service"] 