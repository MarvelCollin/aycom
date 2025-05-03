FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/user
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/user-service .
CMD ["./user-service"] 