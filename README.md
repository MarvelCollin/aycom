# AYCOM Microservices Application

This project is a modern web application built with a microservices architecture.

## Technology Stack

### Frontend
- **Framework**: Svelte with TypeScript
- **Build Tool**: Vite
- **Styling**: Pure CSS/SASS/SCSS (no UI libraries)
- **Design**: Responsive with desktop, tablet, mobile breakpoints

### Backend
- **Language**: Go
- **Architecture**: Microservices
- **Communication**: gRPC between services
- **Messaging**: RabbitMQ for async workflows
- **API Gateway**: For frontend-backend communication
- **Documentation**: Swagger
- **Database**: PostgreSQL (separate for each microservice)
- **Caching**: Redis
- **Logging**: Configurable levels based on environment

### Security
- **Authentication**: JWT with access and refresh tokens
- **Password Security**: Salting and hashing

### Infrastructure
- **Media Storage**: Supabase
- **AI Components**: Flask
- **Containerization**: Docker + Docker Compose
- **Version Control**: Git with structured commits

## Project Structure

```
AYCOM/
├── frontend/                   # Svelte frontend application
├── backend/                    # Go microservices
│   ├── gateway/                # API Gateway
│   ├── event-bus/              # Event Bus for async messaging
│   └── services/               # Microservices
│       ├── auth/               # Authentication service
│       ├── user/               # User management service
│       └── product/            # Product service
├── ai-service/                 # Flask AI service
└── shared/                     # Shared code and protocols
```

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 14 or higher

### Setup and Installation

1. Clone the repository:
   ```
   git clone https://github.com/your-username/aycom.git
   cd aycom
   ```

2. Create a `.env` file based on the example:
   ```
   cp .env.example .env
   ```

3. Start the development environment:
   ```
   docker-compose up
   ```

4. Access the application:
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080
   - Swagger documentation: http://localhost:8080/swagger/index.html
   - RabbitMQ management: http://localhost:15672 (guest/guest)

## Development Guidelines

- Follow the ESLint rules for frontend development
- Use Go idioms and patterns for backend services
- All API endpoints should be documented with Swagger
- Use gRPC for internal service communication
- Implement unit tests for critical functionality

## License

This project is licensed under the MIT License - see the LICENSE file for details. 