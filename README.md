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

## Environment Variables Setup

The application uses environment variables to configure various services and features. These are organized as follows:

### Root `.env` File

The root `.env` file contains global variables used by Docker Compose and shared between services:

```
# Global Configuration
JWT_SECRET=your_jwt_secret

# Google OAuth Credentials
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret

# Database Credentials
AUTH_DB_USER=db_user
AUTH_DB_PASSWORD=db_password
AUTH_DB_NAME=auth_db
...
```

### Frontend `.env` File (frontend/.env)

The frontend `.env` file contains frontend-specific variables, which are prefixed with `VITE_` to make them accessible in the Vite application:

```
# Google OAuth client ID for frontend
VITE_GOOGLE_CLIENT_ID=your_google_client_id

# reCAPTCHA site key
VITE_RECAPTCHA_SITE_KEY=your_recaptcha_site_key

# API base URL
VITE_API_BASE_URL=/api/v1
```

### Backend `.env` File (backend/.env)

The backend `.env` file contains backend-specific variables used by the various microservices:

```
# Service Configuration
API_GATEWAY_PORT=8080
AUTH_SERVICE_ADDR=localhost:50051
...

# Security
JWT_SECRET=your_jwt_secret

# Google OAuth Configuration (for auth service)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
...
```

## Google Authentication Flow

1. The frontend initializes Google Sign-In using the Google Identity Services API
2. When a user clicks the Google Sign-In button, they are redirected to Google for authentication
3. After successful authentication, Google redirects back to the `/google-callback` route
4. The callback page extracts the credential token and sends it to the backend API
5. The backend validates the token with Google and either creates a new user or logs in an existing user
6. The backend returns JWT tokens which are stored by the frontend for authentication

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