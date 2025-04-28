FROM golang:1.21-alpine

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git postgresql-client ca-certificates

# Copy the auth service directory
COPY ./backend/services/auth /app/
COPY ./backend/proto /app/backend/proto

# Initialize the go module if it doesn't already exist
RUN if [ ! -f go.mod ]; then go mod init github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth; fi

# Fix and download dependencies
RUN go mod tidy
RUN go mod vendor

# Build the binary
RUN go build -o auth-service .

# Expose the port
EXPOSE 9090

# Run the binary
CMD ["./auth-service"]