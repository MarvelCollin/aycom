// API Configuration
export const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";

// Default request timeout in milliseconds
export const DEFAULT_TIMEOUT = 30000; // 30 seconds

// Response status codes
export const STATUS_CODES = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
};