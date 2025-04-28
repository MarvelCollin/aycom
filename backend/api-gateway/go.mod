module github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway

go 1.20

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.8.4
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.2
	google.golang.org/grpc v1.55.0
)

// Add replace directive to fix the import path issue
replace github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto/auth => ../services/auth/proto/auth