FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates

# Copy the event-bus directory
COPY ./backend/event-bus /app/

# Download dependencies
RUN go mod download

# Build the binary
RUN go build -o event-bus .

# Final stage: create a clean image
FROM alpine:latest

WORKDIR /app

# Copy necessary certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --from=builder /app/event-bus .

# Set the entrypoint to run the binary
ENTRYPOINT ["./event-bus"] 