# API Gateway Documentation

## Standard API Response Format

All API responses follow a standardized JSON format for consistency:

### Success Response Structure

```json
{
  "success": true,
  "data": {
    // Response data goes here
  },
  "meta": {
    // Optional metadata like pagination info
  }
}
```

### Error Response Structure

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  },
  "meta": {
    // Optional additional error details
  }
}
```

### Field Naming Conventions

All response fields should use `snake_case` format. Here are some standard field names:

- User properties: `id`, `user_id`, `name`, `username`, `email`, `profile_picture_url`, `banner_url`, `bio`, `is_verified`, `is_admin`, `follower_count`, `following_count`, `created_at`, `updated_at`

- Thread properties: `id`, `thread_id`, `content`, `created_at`, `updated_at`, `likes_count`, `replies_count`, `reposts_count`, `views_count`, `is_liked`, `is_reposted`, `is_bookmarked`, `is_pinned`

- Community properties: `id`, `community_id`, `name`, `description`, `logo_url`, `banner_url`, `creator_id`, `is_approved`, `categories`, `created_at`, `member_count`

- Media properties: `id`, `url`, `type`, `alt`

## How to Use Response Utilities

The `utils` package provides standardized response functions:

```go
// Success response
utils.SendSuccessResponse(c, http.StatusOK, data)

// Success response with metadata
utils.SendSuccessResponse(c, http.StatusOK, data, metadata)

// Error response
utils.SendErrorResponse(c, httpStatus, "ERROR_CODE", "Error message")

// Paginated response
utils.SendPaginatedResponse(c, http.StatusOK, items, currentPage, totalPages, totalItems, perPage)

// Validation error response
utils.SendValidationErrorResponse(c, fieldErrors)
``` 