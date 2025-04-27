FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy Go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service .

# Use a smaller image for the final application
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/product-service .

# Set the entrypoint
ENTRYPOINT ["./product-service"] 