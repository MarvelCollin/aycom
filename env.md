# Environment Variables for AYCOM

This document describes the environment variables used in the AYCOM project.

## Core Environment Variables

```env
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=3600
REFRESH_TOKEN_EXPIRY=604800
ENVIRONMENT=development

# API Gateway Configuration
API_GATEWAY_PORT=8083

# Service Ports
AUTH_SERVICE_PORT=50051
USER_SERVICE_PORT=50052
PRODUCT_SERVICE_PORT=50053
THREAD_SERVICE_PORT=9092
COMMUNITY_SERVICE_PORT=9093

# Service Hosts (optional, defaults to localhost)
USER_SERVICE_HOST=localhost
THREAD_SERVICE_HOST=localhost
COMMUNITY_SERVICE_HOST=localhost

# OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
RECAPTCHA_SECRET_KEY=your_recaptcha_secret_key
RECAPTCHA_SITE_KEY=your_recaptcha_site_key

# Database Configuration
AUTH_DB_USER=dbuser
AUTH_DB_PASSWORD=dbpassword
AUTH_DB_NAME=auth_db

USER_DB_USER=dbuser
USER_DB_PASSWORD=dbpassword
USER_DB_NAME=user_db

PRODUCT_DB_USER=dbuser
PRODUCT_DB_PASSWORD=dbpassword
PRODUCT_DB_NAME=product_db

# Frontend Configuration
VITE_API_BASE_URL=http://localhost:8083/api/v1
```

## How Environment Variables Are Loaded

The API Gateway attempts to load the `.env` file from multiple locations in the following order:

1. Current directory
2. Parent directory (`../.env`)  
3. Root project directory (`../../.env`)
4. Executable directory
5. Executable parent directory
6. Executable root directory
7. Explicit paths including `C:/BINUS/TPA/Web/AYCOM/.env`

If no `.env` file is found, the application will use environment variables from the system.

## Default Values

If environment variables are not set, the application will use default values:

- API Gateway Port: 8083
- User Service: localhost:50052
- Thread Service: localhost:9092
- Community Service: localhost:9093
