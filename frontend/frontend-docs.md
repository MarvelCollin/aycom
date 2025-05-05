# AYCOM Frontend Documentation - Backend Integration Requirements

This document outlines the components in the frontend that currently use mock data but need to be integrated with backend API endpoints. It provides a reference for backend developers to understand what endpoints need to be implemented and what data structures are expected.

## Table of Contents
1. [Explore Page](#explore-page)
2. [Notification Page](#notification-page)
3. [Message Page](#message-page)
4. [Bookmarks Page](#bookmarks-page)
5. [Authentication](#authentication)
6. [Missing API Integrations](#missing-api-integrations)

## Explore Page

The Explore page (`src/pages/Explore.svelte`) allows users to search for content, view trending topics, and discover other users and communities.

### Current Mock Implementations

| Feature | Current Implementation | Required Backend |
|---------|------------------------|-----------------|
| Trending Topics | Uses mock data in `getTrends()` in `trends.ts` | Real-time trending hashtags API |
| Search | Uses mock handlers in `searchUsers()`, `searchThreads()`, etc. | Search API with filtering capabilities |
| User Recommendations | Mock data for profile suggestions | User search/recommendation API |

### Required Backend Endpoints

1. **Trends API**
   - `GET /api/trends` - Fetch trending topics/hashtags
   - Already partially implemented in `src/api/trends.ts`
   - Response format:
     ```typescript
     {
       trends: Array<{
         id: string,
         category: string;
         title: string; 
         postCount: string;
       }>
     }
     ```

2. **Search API**
   - `GET /api/users/search` - Search for users, with query params (already implemented in `user.ts`)
   - `GET /api/threads/search` - Search for threads/posts
   - `GET /api/media/search` - Search for media content
   - `GET /api/communities/search` - Search for communities
   - Query parameters: `q` (search term), `filter`, `page`, `limit`, `sort`, etc.

3. **User Recommendations API**
   - `GET /api/users/recommendations` - Get recommended users to follow

### Backend Integration Notes

- The Explore page needs pagination support for all search results
- The component expects real-time updates for trending topics
- Authentication is required for most functionality
- Support for filtering by various categories (all, following, verified)

## Notification Page

The Notification page (`src/pages/Notification.svelte`) displays notifications for user interactions like likes, reposts, follows, and mentions.

### Current Mock Implementations

| Feature | Current Implementation | Required Backend |
|---------|------------------------|-----------------|
| All Notifications | Mock data in `fetchNotifications()` | Real notifications API with filters |
| Mentions | Static mock data | API for mentions with thread context |
| Real-time Updates | Commented mock WebSocket code | WebSocket or polling for real-time notifications |
| Notification Status | Client-side state only | Server persistence of read/unread status |

### Required Backend Endpoints

1. **Notifications API**
   - `GET /api/notifications` - Fetch all notifications for the current user
   - `GET /api/notifications/mentions` - Fetch only mention notifications
   - `PUT /api/notifications/:id/read` - Mark a notification as read
   - `PUT /api/notifications/read-all` - Mark all notifications as read
   - `DELETE /api/notifications/:id` - Delete a notification

2. **Real-time Notifications**
   - WebSocket endpoint: `wss://api.example.com/notifications`
   - Event format:
     ```typescript
     {
       id: string;
       type: 'like' | 'repost' | 'follow' | 'mention';
       userId: string;
       username: string;
       displayName: string;
       avatar: string | null;
       timestamp: string;
       threadId?: string;
       threadContent?: string;
       isRead: boolean;
     }
     ```

3. **Email Notifications**
   - Email service integration (SendGrid, Amazon SES, etc.)
   - Email templates for different notification types
   - User preferences API for notification settings

### Backend Integration Notes

- Notifications must be delivered in real-time (WebSockets preferred)
- Each notification should trigger an email (with user preference control)
- Click handling needs to update read status on the server
- Thread and profile linking requires proper routing with IDs

## Message Page

The Message page (`src/pages/Message.svelte`) allows users to send direct messages to individuals or groups, providing real-time chat functionality.

### Current Mock Implementations

| Feature | Current Implementation | Required Backend |
|---------|------------------------|-----------------|
| Chat List | Uses mock data via `generateMockChats()` instead of `listChats()` from `chat.ts` | Real-time chat list API |
| Chat Messages | Mock data for individual messages | Message history API with pagination |
| Message Sending | Client-side only, no call to `sendMessage()` API | Message sending API with real-time delivery |
| Media Sharing | UI placeholders only | File upload API and media storage service |
| Group Chats | Static mock data | Group management API (create, edit, join, leave) |
| Message Unsend | Client-side only (1-minute window) | Server-side message deletion with timestamp validation |
| Real-time Updates | No implementation | WebSocket for real-time message delivery |

### Required Backend Endpoints

1. **Chats API**
   - `GET /api/chats` - Fetch all chats for the current user (implemented in `chat.ts`)
   - `GET /api/chats/:chatId` - Fetch a specific chat
   - `POST /api/chats` - Create a new chat (implemented in `chat.ts`)
   - `DELETE /api/chats/:chatId` - Delete a chat from user's history
   - Response format matches the implementation in `chat.ts`

2. **Messages API**
   - `GET /api/chats/:chatId/messages` - Fetch messages for a chat (with pagination) (implemented in `chat.ts`)
   - `POST /api/chats/:chatId/messages` - Send a new message (implemented in `chat.ts`)
   - `DELETE /api/chats/:chatId/messages/:messageId` - Delete/unsend a message (implemented in `chat.ts`)
   - `POST /api/chats/:chatId/messages/:messageId/unsend` - Special endpoint for unsending messages (implemented in `chat.ts`)
   - `GET /api/chats/:chatId/messages/search` - Search messages in a chat (implemented in `chat.ts`)

3. **Group Management API**
   - `GET /api/chats/:chatId/participants` - Get participants in a chat (implemented in `chat.ts`)
   - `POST /api/chats/:chatId/participants` - Add participants to a group chat (implemented in `chat.ts`)
   - `DELETE /api/chats/:chatId/participants/:userId` - Remove a participant from a group chat (implemented in `chat.ts`)
   - `PUT /api/chats/:chatId` - Update group chat details (name, avatar)

4. **Media Upload API**
   - `POST /api/media` - Upload a media file (multipart/form-data)
   - Response:
     ```typescript
     {
       id: string;
       type: 'image' | 'gif' | 'video';
       url: string;
       thumbnail?: string;
     }
     ```

5. **Real-time Messaging**
   - WebSocket endpoint: `wss://api.example.com/chat`
   - Events:
     - `message` - New message received
     - `message_update` - Message updated or deleted
     - `typing` - User is typing notification
     - `read_receipt` - Message read by recipient
     - `participant_join` - User joined a group
     - `participant_leave` - User left a group

### Backend Integration Notes

- The messaging system requires low-latency delivery of messages
- Unread count should be maintained on the server and incremented for offline users
- Message history should persist indefinitely or according to data retention policy
- Media files need to be stored securely with appropriate access controls
- Unsending messages must be limited to messages sent within the last minute
- Real-time delivery should handle offline/online status and message queueing
- Search should support finding both chats and messages by content
- Group chat member management requires permission checks (who can add/remove)

## Bookmarks Page

The Bookmarks page (`src/pages/Bookmarks.svelte`) allows users to view and search through all their bookmarked threads.

### Current Mock Implementations

| Feature | Current Implementation | Required Backend |
|---------|------------------------|-----------------|
| Bookmarks List | Mock data in `fetchBookmarks()` | Bookmarks API with user-specific data |
| Search | Client-side filtering with `filterBookmarks()` | Server-side search API (optional) |
| Remove Bookmark | Client-side state update in `handleRemoveBookmark()` | Bookmark removal API endpoint |

### Required Backend Endpoints

1. **Bookmarks API**
   - `GET /api/bookmarks` - Fetch all bookmarks for the current user
   - `DELETE /api/bookmarks/:tweetId` - Remove a bookmark
   - `POST /api/bookmarks/:tweetId` - Add a bookmark (for the TweetCard component)
   - Response format for bookmarks list:
     ```typescript
     {
       bookmarks: Array<ITweet>
     }
     ```

2. **Search API (Optional)**
   - `GET /api/bookmarks/search?q=query` - Search through user's bookmarks
   - Query parameters: `q` (search term)

### Backend Integration Notes

- The Bookmarks page should only display threads bookmarked by the authenticated user
- Authentication is required to access bookmarks
- Client-side search is already implemented, but server-side search could be more efficient for users with many bookmarks
- The bookmark status should be consistent across the application (Feed, Thread Detail, Bookmarks page)

## Authentication

Authentication is implemented across all pages and is partially integrated with existing API endpoints in `src/api/auth.ts`.

### Current Authentication Implementation

| Feature | Current Implementation | Required Backend |
|---------|------------------------|-----------------|
| Login | `login()` in `auth.ts` | User authentication with JWT |
| Registration | `register()` in `auth.ts` | User creation and validation |
| Token Refresh | `refreshToken()` in `auth.ts` | JWT refresh mechanism |
| Email Verification | `verifyEmail()` in `auth.ts` | Email verification system |
| Social Login | `googleLogin()` in `auth.ts` | OAuth integration with Google |

### Required Backend Endpoints

1. **Authentication API**
   - `POST /api/users/login` - Authenticate user and generate tokens (implemented)
   - `POST /api/users/register` - Register a new user (implemented)
   - `POST /api/auth/refresh-token` - Refresh access token using refresh token (implemented)
   - `POST /api/auth/verify-email` - Verify a user's email address (implemented)
   - `POST /api/auth/resend-verification` - Resend verification email (implemented)
   - `POST /api/auth/google` - Authenticate with Google (implemented)

### Authentication Flow

The current authentication implementation uses:

```typescript
function checkAuth() {
  if (!authState.isAuthenticated) {
    toastStore.showToast('You need to log in to access...', 'warning');
    window.location.href = '/login';
    return false;
  }
  return true;
}
```

The backend should implement proper JWT validation and session management to support this authentication flow. Token handling is already implemented with methods for storing and retrieving tokens.

## Missing API Integrations

After reviewing the code, here are the specific missing API integrations that need to be implemented:

### 1. Message Page API Integration

The Message page needs to be updated to use the existing chat APIs:

- Replace `generateMockChats()` with a call to `listChats()` from `../api/chat.ts`
- Implement `sendMessage()` to call the API function of the same name from `../api/chat.ts`
- Integrate `unsendMessage()` with the corresponding API call
- Add functionality for `deleteMessage()`, `searchMessages()`, etc.

### 2. Notification API Endpoints

Currently, there are no API endpoints for notifications. The following need to be created:

- Create a `notification.ts` file in the `api` directory
- Implement API functions for fetching, marking as read, and deleting notifications
- Add WebSocket integration for real-time notifications

### 3. Real-time Integration

For both the Message and Notification pages, WebSocket integration is needed:

- Create a WebSocket utility for real-time communication
- Implement event listeners for message and notification events
- Handle reconnection and authentication over WebSocket

### 4. Explore Page Search Integration

While the API functions exist, they need to be fully integrated:

- Implement `executeSearch()` in Explore.svelte to use the appropriate search functions based on the selected tab
- Add pagination support for search results
- Connect filter changes to search API parameters

### 5. Media Upload API

There's no implementation for media uploads:

- Create a media upload API endpoint
- Implement functions for handling image, GIF, and video uploads
- Add support for displaying uploaded media in messages and threads

### 6. Bookmarks API

There's no implementation for bookmarks API:

- Create a `bookmarks.ts` file in the `api` directory
- Implement functions for fetching, adding, and removing bookmarks
- Ensure bookmarks state is consistent across components that can bookmark threads

## Data Models

### Notification Model

```typescript
interface Notification {
  id: string;
  type: 'like' | 'repost' | 'follow' | 'mention';
  userId: string;
  username: string;
  displayName: string;
  avatar: string | null;
  timestamp: string;
  threadId?: string;
  threadContent?: string;
  isRead: boolean;
}
```

### Chat Models

```typescript
interface Chat {
  id: string;
  type: 'individual' | 'group';
  name: string;
  avatar: string | null;
  participants: Participant[];
  lastMessage?: {
    content: string;
    timestamp: string;
    senderId: string;
  };
  unreadCount: number;
}

interface Message {
  id: string;
  chatId: string;
  senderId: string;
  senderName: string;
  senderAvatar: string | null;
  content: string;
  timestamp: string;
  isDeleted: boolean;
  attachments: Attachment[];
}
```

### User Model

Based on the implementation in `user.ts`:

```typescript
interface User {
  id: string;
  username: string;
  name: string;
  profile_picture_url: string | null;
  bio: string;
  is_verified: boolean;
  follower_count: number;
  is_following: boolean;
}
```

### Bookmark Model

```typescript
interface Bookmark {
  userId: string;
  tweetId: string;
  createdAt: string;
}
```

## Implementation Priority

1. Authentication and authorization (partially implemented)
2. Basic notifications API without real-time updates
3. Search API endpoints (partially implemented)
4. Basic messaging API (send/receive without real-time) (partially implemented)
5. Bookmarks API
6. Real-time notification delivery via WebSockets
7. Real-time messaging via WebSockets
8. Media upload and storage
9. Group chat management (partially implemented)
10. Email notification integration (partially implemented)
11. Trending topics API (partially implemented)
