FROM golang:1.24.2-alpine

WORKDIR /app

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

COPY go.mod go.sum ./

RUN mkdir -p services/thread services/user/proto

WORKDIR /app/services/thread

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

ENV PORT=9092
ENV HEALTH_PORT=8082
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

EXPOSE 9092 8082

CMD ["air", "-c", ".air.toml"] 