FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN mkdir -p api-gateway

WORKDIR /app/api-gateway

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

ENV GIN_MODE=debug
ENV PORT=8081
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

EXPOSE 8081

CMD ["air", "-c", ".air.toml"]