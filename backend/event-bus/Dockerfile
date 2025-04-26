FROM golang:1.20-alpine

WORKDIR /app

# Create a simple executable
RUN echo 'package main' > main.go && \
    echo 'import "fmt"' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '    fmt.Println("Event bus service started")' >> main.go && \
    echo '    select{}' >> main.go && \
    echo '}' >> main.go

# Initialize go module
RUN go mod init event-bus

# Build the executable
RUN go build -o event-bus .

# Use a smaller image for the final application
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=0 /app/event-bus .

# Set the entrypoint
ENTRYPOINT ["./event-bus"] 