# Builder stage
FROM golang:1.24-alpine AS builder

# Install only necessary build tools (git for dependencies)
RUN apk add --no-cache git

WORKDIR /go/src/aycom/backend

COPY go.mod go.sum ./

COPY . .

RUN go mod download

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api-gateway-app ./api-gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /app/api-gateway-app /app/app

ENV GIN_MODE=debug

USER appuser

EXPOSE 8081

CMD ["./app"]