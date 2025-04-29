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

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --chown=appuser:appgroup --from=builder /app/event-bus .

# Switch to the non-root user
USER appuser

# Set the entrypoint to run the binary
ENTRYPOINT ["./event-bus"] 