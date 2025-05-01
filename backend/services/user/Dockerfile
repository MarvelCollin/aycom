FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy the user service files
COPY . .

# Create a setup script
RUN echo '#!/bin/sh' > /app/setup.sh && \
    echo 'mkdir -p /app/proto' >> /app/setup.sh && \
    echo 'echo "module github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto" > /app/proto/go.mod' >> /app/setup.sh && \
    echo 'echo "go 1.21" >> /app/proto/go.mod' >> /app/setup.sh && \
    echo 'echo "module github.com/Acad600-Tpa/WEB-MV-242/backend/services/user" > /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "go 1.21" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "require (" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "  github.com/google/uuid v1.6.0" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "  google.golang.org/grpc v1.72.0" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "  gorm.io/driver/postgres v1.5.2" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "  gorm.io/gorm v1.25.2" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo ")" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "replace (" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo "  github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto => ./proto" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'echo ")" >> /app/go.mod.tmp' >> /app/setup.sh && \
    echo 'mv /app/go.mod.tmp /app/go.mod' >> /app/setup.sh && \
    echo 'mkdir -p /app/pkg/db' >> /app/setup.sh && \
    echo 'if [ ! -s /app/pkg/db/database.go ]; then' >> /app/setup.sh && \
    echo '  echo "package db" > /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "import (" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "  \"fmt\"" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "  \"gorm.io/driver/postgres\"" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "  \"gorm.io/gorm\"" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo ")" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "// NewDatabase creates a new database connection" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "func NewDatabase(dsn string) (*gorm.DB, error) {" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "  return gorm.Open(postgres.Open(dsn), &gorm.Config{})" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo '  echo "}" >> /app/pkg/db/database.go' >> /app/setup.sh && \
    echo 'fi' >> /app/setup.sh && \
    echo 'find . -name "*_test.go" -size 0 -exec sh -c "echo \"package \$(basename \$(dirname {}))\" > {}" \;' >> /app/setup.sh && \
    echo 'go mod tidy' >> /app/setup.sh && \
    chmod +x /app/setup.sh

# Run the setup script
RUN /app/setup.sh

# Build the service
RUN go build -o user-service main.go || exit 0

# Final stage
FROM alpine:latest

WORKDIR /app

# Add basic utilities and certificates
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from the builder stage
COPY --from=builder /app/user-service .

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Make app directory writable for the non-root user
RUN chown -R appuser:appgroup /app

# Use non-root user
USER appuser

# Expose the service port
EXPOSE 9091

# Run the service
CMD ["./entrypoint.sh"]