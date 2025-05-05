# AYCOM Backend Project Tasks & Reminders

## Infrastructure Reminders
- **Docker Compose**: Configuration is located outside of this project. Remember to maintain synchronization between service definitions and the external docker-compose.yml file
- **Environment Variables**: Each service has its own .env file that must be properly configured

## Dependencies
- Swagger dependencies missing in go.mod:
  ```
  github.com/swaggo/files
  github.com/swaggo/gin-swagger
  github.com/swaggo/swag
  ```
  Run: `go get github.com/swaggo/files github.com/swaggo/gin-swagger github.com/swaggo/swag`

## Service Implementation Status
- **API Gateway**: Routes structure ready, Swagger integration needs fixing
- **Event Bus**: Basic implementation exists, but needs complete RabbitMQ connection handling
- **User Service**: Authentication handlers need real implementations
- **Community Service**: Routes need path fixes (add leading slashes)
- **Thread Service**: Core functionality needs implementation

## Next Steps
- Fix Swagger integration to enable API documentation
- Complete real implementations of mock services
- Enhance database connection handling with proper pooling and retry logic
- Implement proper error handling across microservices
- Standardize environment variable loading across services

## Testing
- Create integration tests between services
- Implement end-to-end testing with Docker Compose environment
