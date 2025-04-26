FROM golang:1.20-alpine

WORKDIR /app

# Create a simple executable
RUN echo 'package main' > main.go && \
    echo 'import "fmt"' >> main.go && \
    echo 'import "os"' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '    fmt.Println("Auth service started on port:", os.Getenv("AUTH_SERVICE_PORT"))' >> main.go && \
    echo '    select{}' >> main.go && \
    echo '}' >> main.go

# Initialize go module
RUN go mod init auth-service

# Build the executable
RUN go build -o auth-service .

# Use a smaller image for the final application
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=0 /app/auth-service .

# Set the entrypoint
ENTRYPOINT ["./auth-service"] 