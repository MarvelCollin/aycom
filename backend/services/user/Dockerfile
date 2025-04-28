FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod tidy
RUN apk add --no-cache git

# Copy Go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Run go mod tidy to ensure dependencies are correct
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service .

# Use a smaller image for the final application
FROM alpine:3.19

# Install PostgreSQL client and CA certificates
RUN apk --no-cache add ca-certificates postgresql-client && \
    # Create a non-root user to run the application
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/user-service .

# Use the non-root user
USER appuser

EXPOSE 9091

CMD ["./user-service"]