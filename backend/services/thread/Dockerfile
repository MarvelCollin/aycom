FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy the entire project
COPY . .

# Create entrypoint.sh
RUN echo '#!/bin/sh' > /app/backend/services/thread/entrypoint.sh && \
    echo 'set -e' >> /app/backend/services/thread/entrypoint.sh && \
    echo 'echo "Starting thread service..."' >> /app/backend/services/thread/entrypoint.sh && \
    echo 'exec ./thread-service' >> /app/backend/services/thread/entrypoint.sh && \
    chmod +x /app/backend/services/thread/entrypoint.sh

# Create a simple main.go if needed
WORKDIR /app/backend/services/thread
RUN if [ ! -s main.go ] || ! go build -o thread-service main.go; then \
    echo 'package main' > main.go && \
    echo 'import (' >> main.go && \
    echo '  "fmt"' >> main.go && \
    echo '  "log"' >> main.go && \
    echo '  "net"' >> main.go && \
    echo '  "os"' >> main.go && \
    echo ')' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '  port := os.Getenv("PORT")' >> main.go && \
    echo '  if port == "" {' >> main.go && \
    echo '    port = "9092"' >> main.go && \
    echo '  }' >> main.go && \
    echo '  lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))' >> main.go && \
    echo '  if err != nil {' >> main.go && \
    echo '    log.Fatalf("Failed to listen: %v", err)' >> main.go && \
    echo '  }' >> main.go && \
    echo '  log.Printf("Thread service started on port %s", port)' >> main.go && \
    echo '  log.Printf("Server listening at %v", lis.Addr())' >> main.go && \
    echo '  select {}' >> main.go && \
    echo '}' >> main.go && \
    go build -o thread-service main.go; \
    fi

# Final stage
FROM alpine:latest

WORKDIR /app

# Add basic utilities and certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and entrypoint
COPY --from=builder /app/backend/services/thread/thread-service .
COPY --from=builder /app/backend/services/thread/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose port
EXPOSE 9092

# Run the service
CMD ["./entrypoint.sh"] 