FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Create a minimal working program
RUN printf 'package main\n\nimport (\n    "fmt"\n    "log"\n    "net/http"\n)\n\nfunc main() {\n    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {\n        fmt.Fprintf(w, "Auth Service - Development Placeholder")\n    })\n\n    log.Println("Auth Service starting on :9090")\n    log.Fatal(http.ListenAndServe(":9090", nil))\n}' > main.go

# Build a static binary
RUN CGO_ENABLED=0 go build -o auth-service main.go

# Use a minimal alpine image for the final container
FROM alpine:latest

# Add basic utilities and certificates
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Use non-root user
USER appuser

# Expose the service port
EXPOSE 9090

# Run the service
CMD ["./auth-service"]