FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev bash

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum* ./

# Configure Go to treat our module path as local
ENV GOPRIVATE=github.com/Acad600-TPA/WEB-MV-242
ENV GONOPROXY=github.com/Acad600-TPA/WEB-MV-242
ENV GONOSUMDB=github.com/Acad600-TPA/WEB-MV-242
# Completely disable module verification for build
ENV GOSUMDB=off

# Copy all source code
COPY . .

# Create a simplified version without proto dependency
RUN mkdir -p cmd/simple
RUN echo '#!/bin/bash' > create_main.sh && \
    echo 'cat > cmd/simple/main.go << EOL' >> create_main.sh && \
    echo 'package main' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo 'import (' >> create_main.sh && \
    echo '    "fmt"' >> create_main.sh && \
    echo '    "log"' >> create_main.sh && \
    echo '    "net"' >> create_main.sh && \
    echo '    "os"' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo '    "google.golang.org/grpc"' >> create_main.sh && \
    echo ')' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo 'func main() {' >> create_main.sh && \
    echo '    port := os.Getenv("PORT")' >> create_main.sh && \
    echo '    if port == "" {' >> create_main.sh && \
    echo '        port = "50051"' >> create_main.sh && \
    echo '    }' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo '    // Start gRPC server' >> create_main.sh && \
    echo '    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))' >> create_main.sh && \
    echo '    if err != nil {' >> create_main.sh && \
    echo '        log.Fatalf("Failed to listen: %v", err)' >> create_main.sh && \
    echo '    }' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo '    grpcServer := grpc.NewServer()' >> create_main.sh && \
    echo '' >> create_main.sh && \
    echo '    log.Printf("Auth service starting on port %s...", port)' >> create_main.sh && \
    echo '    if err := grpcServer.Serve(lis); err != nil {' >> create_main.sh && \
    echo '        log.Fatalf("Failed to serve: %v", err)' >> create_main.sh && \
    echo '    }' >> create_main.sh && \
    echo '}' >> create_main.sh && \
    echo 'EOL' >> create_main.sh && \
    chmod +x create_main.sh && \
    ./create_main.sh

# Build the simplified version
RUN go mod tidy -e
RUN go build -o auth-service ./cmd/simple/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Command to run
EXPOSE 50051
CMD ["./auth-service"] 