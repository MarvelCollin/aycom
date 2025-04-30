FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Create a minimal working program with CORS support
RUN printf 'package main\n\nimport (\n    "fmt"\n    "log"\n    "net/http"\n)\n\n// CORS middleware\nfunc corsMiddleware(next http.Handler) http.Handler {\n    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n        // Set CORS headers\n        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")\n        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")\n        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")\n        w.Header().Set("Access-Control-Allow-Credentials", "true")\n\n        // Handle preflight requests\n        if r.Method == "OPTIONS" {\n            w.WriteHeader(http.StatusOK)\n            return\n        }\n\n        next.ServeHTTP(w, r)\n    })\n}\n\nfunc main() {\n    // Define routes\n    mux := http.NewServeMux()\n    \n    // API routes\n    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {\n        fmt.Fprintf(w, "API Gateway - Development Placeholder")\n    })\n\n    // Auth routes\n    mux.HandleFunc("/api/v1/auth/google", func(w http.ResponseWriter, r *http.Request) {\n        w.Header().Set("Content-Type", "application/json")\n        fmt.Fprintf(w, `{"success":true,"message":"Development placeholder for Google Auth"}`)\n    })\n\n    // Wrap with CORS middleware\n    handler := corsMiddleware(mux)\n    \n    log.Println("API Gateway starting on :8080")\n    log.Fatal(http.ListenAndServe(":8080", handler))\n}' > main.go

# Build a static binary
RUN CGO_ENABLED=0 go build -o api-gateway main.go

# Use a minimal alpine image for the final container
FROM alpine:latest

# Add basic utilities and certificates
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/api-gateway .

# Use non-root user
USER appuser

# Expose the service port
EXPOSE 8080

# Run the service
CMD ["./api-gateway"] 