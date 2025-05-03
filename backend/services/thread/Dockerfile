FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/thread
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/thread-service ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/thread-service .
CMD ["./thread-service"] 