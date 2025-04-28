FROM golang:1.21-alpine

WORKDIR /app

# Install dependencies
RUN apk --no-cache add git

# Copy the API Gateway code
COPY ./backend/api-gateway /app/

# Create a minimal main.go file for testing
RUN echo 'package main' > /app/main-minimal.go && \
    echo 'import (' >> /app/main-minimal.go && \
    echo '    "net/http"' >> /app/main-minimal.go && \
    echo '    "github.com/gin-gonic/gin"' >> /app/main-minimal.go && \
    echo ')' >> /app/main-minimal.go && \
    echo 'func main() {' >> /app/main-minimal.go && \
    echo '    r := gin.Default()' >> /app/main-minimal.go && \
    echo '    r.GET("/healthcheck", func(c *gin.Context) {' >> /app/main-minimal.go && \
    echo '        c.JSON(http.StatusOK, gin.H{' >> /app/main-minimal.go && \
    echo '            "status": "UP",' >> /app/main-minimal.go && \
    echo '            "message": "API Gateway is running",' >> /app/main-minimal.go && \
    echo '        })' >> /app/main-minimal.go && \
    echo '    })' >> /app/main-minimal.go && \
    echo '    r.Run(":8080")' >> /app/main-minimal.go && \
    echo '}' >> /app/main-minimal.go

# Initialize or update go.mod
RUN if [ ! -f go.mod ]; then go mod init github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway; fi

# Download required packages
RUN go get github.com/gin-gonic/gin

# Build minimal version for testing
RUN go build -o api-gateway main-minimal.go

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-gateway"] 