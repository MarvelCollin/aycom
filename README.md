# AYCOM - Microservices Application

AYCOM is a full-stack microservices application with a Svelte frontend and Go microservices backend.

## Architecture

The application consists of the following components:

- **Frontend**: Svelte application
- **API Gateway**: Go-based API Gateway with Swagger documentation
- **Auth Service**: Handles authentication and authorization
- **User Service**: Manages user profiles and data
- **Product Service**: Manages product data
- **Event Bus**: Handles event-driven communication between services
- **AI Service**: Provides AI-related functionality

## Prerequisites

- Docker and Docker Compose
- Git

## Getting Started

1. Clone the repository:
   ```
   git clone https://github.com/your-username/AYCOM.git
   cd AYCOM
   ```

2. Start the application:
   ```
   docker-compose up -d
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - API Gateway Swagger Documentation: http://localhost:8080/swagger/index.html
   - RabbitMQ Management: http://localhost:15672 (username: guest, password: guest)

## Service Endpoints

- Frontend: `http://localhost:3000`
- API Gateway: `http://localhost:8080`
- Auth Service (gRPC): `localhost:50051`
- User Service (gRPC): `localhost:50052`
- Product Service (gRPC): `localhost:50053`
- AI Service: `http://localhost:5000`

## Databases

- Auth Database: PostgreSQL on port 5432
- User Database: PostgreSQL on port 5433
- Product Database: PostgreSQL on port 5434
- Redis: Redis on port 6379

## Development

### Building and running individual services

You can build and run individual services using Docker:

```bash
# Build and run the frontend
cd frontend
docker build -t aycom-frontend .
docker run -p 3000:3000 aycom-frontend

# Build and run the gateway
cd backend/gateway
docker build -t aycom-gateway .
docker run -p 8080:8080 aycom-gateway
```

### Environment Variables

Each service uses environment variables for configuration. See the docker-compose.yml file for details.

## Troubleshooting

- **Crypto Issues**: If you encounter crypto-related errors in the frontend, the environment variable `NODE_OPTIONS=--openssl-legacy-provider` is included to address this.
- **Database Connection Issues**: Ensure the database containers are running. You can connect to them using a PostgreSQL client.
- **Service Communication**: The services are configured to communicate with each other using the service names defined in docker-compose.yml.

## License

MIT 