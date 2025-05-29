# API Naming Conventions

To ensure consistency across the API, the following naming conventions should be followed:

## General Guidelines

1. All field names in API responses should use `snake_case` format
2. All error codes should use `UPPER_SNAKE_CASE` format
3. All boolean fields should be prefixed with `is_`, `has_`, or `can_` as appropriate

## Standard Field Names

### User Properties
- `id` - Unique identifier for the user
- `user_id` - When referencing a user in another object
- `name` - User's display name
- `username` - User's unique username
- `email` - User's email address
- `profile_picture_url` - URL to user's profile picture
- `banner_url` - URL to user's banner image
- `bio` - User's biography/description
- `is_verified` - Whether the user is verified
- `is_admin` - Whether the user is an admin
- `follower_count` - Number of followers
- `following_count` - Number of users being followed
- `created_at` - When the user was created
- `updated_at` - When the user was last updated

### Thread Properties
- `id` - Unique identifier for the thread
- `thread_id` - When referencing a thread in another object
- `content` - Thread text content
- `created_at` - When the thread was created
- `updated_at` - When the thread was last updated
- `likes_count` - Number of likes
- `replies_count` - Number of replies
- `reposts_count` - Number of reposts
- `views_count` - Number of views
- `is_liked` - Whether the current user has liked the thread
- `is_reposted` - Whether the current user has reposted the thread
- `is_bookmarked` - Whether the current user has bookmarked the thread
- `is_pinned` - Whether the thread is pinned

### Community Properties
- `id` - Unique identifier for the community
- `community_id` - When referencing a community in another object
- `name` - Community name
- `description` - Community description
- `logo_url` - URL to community logo
- `banner_url` - URL to community banner
- `creator_id` - ID of the user who created the community
- `is_approved` - Whether the community is approved
- `categories` - List of categories the community belongs to
- `created_at` - When the community was created
- `member_count` - Number of members in the community

### Media Properties
- `id` - Unique identifier for the media
- `url` - URL to the media file
- `type` - Type of media (image, video, etc.)
- `alt` - Alternative text for the media

### Pagination Properties
- `total_count` - Total number of items available
- `current_page` - Current page number
- `per_page` - Number of items per page
- `has_more` - Boolean indicating if there are more pages
- `total_pages` - Total number of pages available

## Using Response Utilities

Always use the provided utility functions for sending responses:

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

This ensures that all responses follow the standard format and structure. 

## Standard Response Formats

### Paginated Response Format
When returning paginated results, use the following format:

```json
{
  "success": true,
  "data": {
    "items": [...],  // The actual items (users, threads, communities, etc.)
    "pagination": {
      "total_count": 100,
      "current_page": 2,
      "per_page": 10,
      "has_more": true,
      "total_pages": 10
    }
  }
}
``` 