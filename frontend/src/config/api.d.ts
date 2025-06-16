// Type declarations for the API configuration
declare module "../config/api" {
  export const API_URL: string;
  export const DEFAULT_TIMEOUT: number;
  export const STATUS_CODES: {
    OK: number;
    CREATED: number;
    BAD_REQUEST: number;
    UNAUTHORIZED: number;
    FORBIDDEN: number;
    NOT_FOUND: number;
    INTERNAL_SERVER_ERROR: number;
  };
}