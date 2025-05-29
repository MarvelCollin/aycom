# API Field Naming Conventions

This document outlines the field naming conventions used in the AYCOM application for API requests and responses.

## Standard Convention

- All API response field names should use `snake_case` format.
- All request payload field names should also use `snake_case` for consistency.
- Boolean field names should use the `is_` prefix for clarity (e.g., `is_verified`, `is_following`).
- Count fields should use the `_count` suffix (e.g., `follower_count`, `likes_count`).

## Common Field Mappings

When working with API responses, you may need to manually map fields if there are inconsistencies. Use this table as a reference:

| Standard Snake Case | Alternative Names  |
|---------------------|-------------------|
| user_id             | userId, id        |
| name                | displayName, display_name |
| profile_picture_url | avatar, profilePictureUrl |
| banner_url          | bannerUrl, backgroundBanner |
| is_verified         | verified, isVerified |
| follower_count      | followerCount, followers_count |
| created_at          | timestamp, createdAt |
| likes_count         | likes |
| replies_count       | replies |
| is_following        | isFollowing |

## Implementation Approach

For simplicity, we use direct field mapping in components and API functions. Example:

```javascript
// Simple approach to normalize user fields
const user = {
  id: apiUser.id,
  username: apiUser.username,
  name: apiUser.name || apiUser.display_name || '',
  profile_picture_url: apiUser.profile_picture_url || apiUser.avatar || '',
  is_verified: apiUser.is_verified || apiUser.verified || false
};
```

This direct mapping is preferred over complex utilities to make the code easier to understand and maintain. 