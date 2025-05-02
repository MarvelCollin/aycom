FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY ../../../backend/go.mod ../../../backend/go.sum ./
RUN go mod download

COPY ../../../backend ./backend
WORKDIR /app/backend/services/user
RUN go build -o user-service .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/backend/services/user/user-service .
EXPOSE 9091
CMD ["./user-service"] 