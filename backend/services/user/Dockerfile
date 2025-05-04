# Base image for building the application
FROM golang:1.24-alpine AS builder

# Set environment variables for Go tools
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Install necessary tools: git, protobuf, and Go protobuf/gRPC plugins
RUN apk add --no-cache git protobuf protobuf-dev && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy go module files and download dependencies first to leverage Docker cache
# Paths are relative to the build context (./backend)
COPY go.mod go.sum /app/
RUN cd /app && go mod download

# Copy the rest of the application code from the build context (./backend)
COPY . /app

WORKDIR /app/services/user

# Now generate proto code within the build stage
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/user.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./main.go

# --- Final Stage ---
# Use a minimal base image for the final container
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/user-service .

# Expose the port the service runs on
EXPOSE 9091

# Command to run the application
ENTRYPOINT ["/app/user-service"] 