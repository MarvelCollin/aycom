FROM golang:1.24-alpine

WORKDIR /app

# Install necessary tools and Air for hot reloading
RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

# Copy Go module files for dependency caching
COPY go.mod go.sum ./

# Only prepare directories - actual code will be mounted via volume in development
RUN mkdir -p event-bus

# Set up for Air in development mode
WORKDIR /app/event-bus

# Create default Air config
RUN touch .air.toml && \
    echo 'root = "."' > .air.toml && \
    echo 'tmp_dir = "tmp"' >> .air.toml && \
    echo '[build]' >> .air.toml && \
    echo '  cmd = "go build -o ./tmp/main ."' >> .air.toml && \
    echo '  bin = "./tmp/main"' >> .air.toml && \
    echo '  delay = 1000' >> .air.toml && \
    echo '  exclude_dir = ["assets", "tmp", "vendor"]' >> .air.toml && \
    echo '  include_ext = ["go", "tpl", "tmpl", "html"]' >> .air.toml && \
    echo '  kill_delay = "0s"' >> .air.toml && \
    echo '  log = "build-errors.log"' >> .air.toml && \
    echo '  stop_on_error = true' >> .air.toml && \
    echo '[screen]' >> .air.toml && \
    echo '  clear_on_rebuild = true' >> .air.toml

# Set environment variables
ENV PORT=8000
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

# Expose port
EXPOSE 8000

# Use Air for hot reloading in development
CMD ["air", "-c", ".air.toml"] 