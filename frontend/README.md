# AYCOM Frontend

## API Field Naming Conventions

This project uses consistent snake_case field naming for all API interactions to match the backend API. Below are the standardized field names for main data types:

### User Fields
- `id` - User ID
- `username` - User's unique username
- `name` - User's display name
- `bio` - User's biography
- `profile_picture_url` - URL to user's profile picture
- `banner_url` - URL to user's banner image
- `is_verified` - Whether user is verified
- `is_admin` - Whether user is an admin
- `follower_count` - Number of followers
- `following_count` - Number of users being followed
- `created_at` - When user was created
- `is_following` - Whether the current user is following this user

### Thread Fields
- `id` - Thread ID
- `thread_id` - Thread identifier
- `content` - Thread text content
- `created_at` - When the thread was created
- `user_id` - ID of thread creator
- `likes_count` - Number of likes
- `replies_count` - Number of replies
- `reposts_count` - Number of reposts
- `views_count` - Number of views
- `is_liked` - Whether current user has liked the thread
- `is_reposted` - Whether current user has reposted the thread
- `is_bookmarked` - Whether current user has bookmarked the thread
- `is_pinned` - Whether thread is pinned

### Community Fields
- `id` - Community ID
- `name` - Community name
- `description` - Community description
- `logo_url` - URL to community logo
- `banner_url` - URL to community banner
- `creator_id` - ID of community creator
- `is_approved` - Whether community is approved
- `member_count` - Number of community members
- `created_at` - When community was created

### Pagination Fields
- `total_count` - Total number of items available
- `current_page` - Current page number
- `per_page` - Number of items per page
- `has_more` - Whether there are more pages available
- `total_pages` - Total number of pages

### API Normalization

To handle any inconsistencies between backend API responses and frontend models, use the normalization utilities in `src/utils/api-normalization.ts`:

```typescript
import { normalizeUser, normalizeThread, normalizeCommunity, normalizePagination } from '../utils/api-normalization';

// Example usage
const normalizedUser = normalizeUser(apiUserResponse);
const normalizedThread = normalizeThread(apiThreadResponse);
``` 