FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies and Swagger
RUN apk add --no-cache git && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Create a simple executable
RUN echo 'package main' > main.go && \
    echo 'import "fmt"' >> main.go && \
    echo 'import "os"' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '    fmt.Println("User service started on port:", os.Getenv("USER_SERVICE_PORT"))' >> main.go && \
    echo '    select{}' >> main.go && \
    echo '}' >> main.go

# Initialize go module and add required dependencies
RUN go mod init user-service && \
    go get github.com/gin-gonic/gin && \
    go get github.com/swaggo/swag

# Build the executable
RUN go build -o user-service .

# Use a smaller image for the final application
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/user-service .

# Set the entrypoint
ENTRYPOINT ["./user-service"] 