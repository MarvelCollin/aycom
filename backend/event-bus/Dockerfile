FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum* ./

# Configure Go to treat our module path as local
ENV GOPRIVATE=github.com/Acad600-TPA/WEB-MV-242
ENV GONOPROXY=github.com/Acad600-TPA/WEB-MV-242
ENV GONOSUMDB=github.com/Acad600-TPA/WEB-MV-242
# Completely disable module verification for build
ENV GOSUMDB=off

# Copy the entire source code
COPY . .

# Build the application
RUN go mod tidy -e
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o event-bus ./cmd/main.go

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/event-bus .

# Command to run when the container starts
CMD ["./event-bus"] 