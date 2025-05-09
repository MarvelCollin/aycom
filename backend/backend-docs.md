Folder Structure 
api -> like handlers that will contain handlers for the api
db -> repository, db, seeder, migration
models -> database schema with gorm
proto -> grpc proto
service -> contain the services

## API Gateway Endpoints

This section lists all the HTTP endpoints defined in the API Gateway (`backend/api-gateway`).

**Base Path:** `/api/v1`

### Health Check

*   `GET /health`: Checks the health status of the API Gateway itself.

### Public Routes

*   `GET /auth/oauth-config`: Retrieves OAuth configuration details.
*   `POST /users/register`: Handles new user registration.
*   `POST /users/login`: Handles user login.
*   `POST /auth/forgot-password`: Request a password reset by providing email. Returns security question.
*   `POST /auth/verify-security-answer`: Verify security answer for password reset. Returns a reset token.
*   `POST /auth/reset-password`: Reset password using a valid token.

### Protected Routes (Require JWT Authentication)

**Authentication (`/auth`)**

*   `POST /login`: User login (Alternative, potentially deprecated if `/users/login` is primary).
*   `POST /register`: User registration (Alternative, potentially deprecated if `/users/register` is primary).
*   `POST /refresh-token`: Refreshes an expired access token using a valid refresh token.
*   `POST /logout`: Invalidates the user's session/token.
*   *Commented Out:* `/verify-email`, `/resend-verification`, `/google` (These routes exist in handlers but are commented out in `auth_routes.go`)

**Users (`/users`)**

*   `GET /profile`: Retrieves the profile of the authenticated user.
*   `PUT /profile`: Updates the profile of the authenticated user.
*   *Missing:* `GET /suggestions` (Called by frontend but not defined in the router).

**Threads (`/threads`)**

*   `POST /`: Creates a new thread.
*   `GET /:id`: Retrieves a specific thread by its ID.
*   `GET /user/:id`: Retrieves all threads posted by a specific user ID.
*   `PUT /:id`: Updates an existing thread by its ID.
*   `DELETE /:id`: Deletes a thread by its ID.
*   `POST /media`: Handles media uploads for threads.

**Trends (`/trends`)**

*   `GET /`: Retrieves current trending topics.

**Products (`/products`)** (Implemented with basic functionality)

*   `GET /`: Lists available products.
*   `GET /:id`: Retrieves a specific product by its ID.
*   `POST /`: Creates a new product.
*   `PUT /:id`: Updates an existing product by its ID.
*   `DELETE /:id`: Deletes a product by its ID.

**Payments (`/payments`)** (Commented out in `routes.go`)

*   `POST /`: Creates a payment record.
*   `GET /:id`: Retrieves a specific payment record.
*   `GET /history`: Retrieves the payment history for the user.