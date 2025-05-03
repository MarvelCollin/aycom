\
## Frontend API Endpoint Usage

This section lists the backend API endpoints currently called by the frontend API client (`src/api`).

**Base URL:** `http://localhost:8081/api/v1` (Configurable via `VITE_API_BASE_URL`)

### Authentication (`src/api/auth.ts`)

*   `POST /users/login`: User login.
*   `POST /users/register`: User registration.
*   `POST /auth/refresh-token`: Refresh authentication token.
*   `POST /auth/verify-email`: Verify user email (Backend endpoint might be missing/commented out).
*   `POST /auth/resend-verification`: Resend email verification code (Backend endpoint might be missing/commented out).
*   `POST /auth/google`: Initiate Google login flow (Backend endpoint might be missing/commented out).

### User (`src/api/user.ts`)

*   `GET /users/profile`: Fetch the profile of the authenticated user.
*   `PUT /users/profile`: Update the profile of the authenticated user.

### Thread (`src/api/thread.ts`)

*   `POST /threads`: Create a new thread.
*   `GET /threads/:id`: Fetch a specific thread by its ID.
*   `GET /threads/user/:userId`: Fetch all threads created by a specific user.
*   `PUT /threads/:id`: Update an existing thread.
*   `DELETE /threads/:id`: Delete a thread.
*   `POST /threads/media`: Upload media files associated with a thread.

### Trends (`src/api/trends.ts`)

*   `GET /trends`: Fetch trending topics.

### Suggestions (`src/api/suggestions.ts`)

*   `GET /users/suggestions`: Fetch suggested users to follow (Backend endpoint might be missing).
