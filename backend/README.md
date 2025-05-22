# AYCOM Backend Services

## Architecture Overview

The AYCOM backend is composed of several microservices organized in a loosely-coupled architecture:

1. **api-gateway** - Entry point that routes requests to appropriate services
2. **event-bus** - Handles asynchronous communication between services
3. **services** - Core business logic divided into domain services
4. **ai-service** - AI capabilities for content analysis and recommendations

## Service Details

### 1. User Service

Handles user authentication, profiles, and account management.

#### Interfaces

```go
// UserService defines operations for user management
type UserService interface {
    CreateUser(ctx context.Context, req *user.CreateUserRequest) (*model.User, error)
    GetUserByID(ctx context.Context, userID string) (*model.User, error)
    GetUserByUsername(ctx context.Context, username string) (*model.User, error)
    UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error)
    DeleteUser(ctx context.Context, userID string) error
    // Authentication methods
    Login(ctx context.Context, username, password string) (string, string, error)
    RefreshToken(ctx context.Context, refreshToken string) (string, error)
}
```

#### Endpoints

- **POST /users** - Create new user
- **GET /users/:id** - Get user profile
- **GET /users/username/:username** - Get user by username
- **PUT /users/:id** - Update user profile
- **DELETE /users/:id** - Delete user account
- **POST /auth/login** - Login and get tokens
- **POST /auth/refresh** - Refresh access token

### 2. Thread Service

Manages threads (posts), replies, media, and interactions.

#### Interfaces

```go
// ThreadService defines operations for thread management
type ThreadService interface {
    CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*model.Thread, error)
    GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error)
    GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, error)
    GetAllThreads(ctx context.Context, page, limit int) ([]*model.Thread, error)
    UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*model.Thread, error)
    DeleteThread(ctx context.Context, threadID, userID string) error
    GetMediaByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Media, error)
    PinThread(ctx context.Context, threadID, userID string) error
    UnpinThread(ctx context.Context, threadID, userID string) error
}

// CategoryService defines operations for thread categories
type CategoryService interface {
    CreateCategory(ctx context.Context, name string, categoryType string) (*model.Category, error)
    GetCategoryByID(ctx context.Context, categoryID string) (*model.Category, error)
    GetAllCategories(ctx context.Context, categoryType string) ([]*model.Category, error)
    UpdateCategory(ctx context.Context, categoryID string, name string) (*model.Category, error)
    DeleteCategory(ctx context.Context, categoryID string) error
    AddCategoryToThread(ctx context.Context, threadID string, categoryID string) error
    RemoveCategoryFromThread(ctx context.Context, threadID string, categoryID string) error
    GetThreadCategories(ctx context.Context, threadID string) ([]*model.Category, error)
    GetOrCreateCategoriesByNames(ctx context.Context, categoryNames []string, categoryType string) ([]string, error)
}

// InteractionService defines operations for thread/reply interactions
type InteractionService interface {
    LikeThread(ctx context.Context, userID, threadID string) error
    UnlikeThread(ctx context.Context, userID, threadID string) error
    LikeReply(ctx context.Context, userID, replyID string) error
    UnlikeReply(ctx context.Context, userID, replyID string) error
    HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error)
    HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error)
    RepostThread(ctx context.Context, userID, threadID string, repostText *string) error
    RemoveRepost(ctx context.Context, userID, threadID string) error
    HasUserReposted(ctx context.Context, userID, threadID string) (bool, error)
    BookmarkThread(ctx context.Context, userID, threadID string) error
    RemoveBookmark(ctx context.Context, userID, threadID string) error
    HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error)
    GetUserBookmarks(ctx context.Context, userID string, page, limit int) ([]*Thread, int64, error)
    BookmarkReply(ctx context.Context, userID, replyID string) error
    RemoveReplyBookmark(ctx context.Context, userID, replyID string) error
    HasUserBookmarkedReply(ctx context.Context, userID, replyID string) (bool, error)
    GetLikedThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]string, error)
}
```

#### Endpoints

- **POST /threads** - Create new thread
- **GET /threads/:id** - Get thread by ID
- **GET /threads** - Get all threads with pagination
- **GET /threads/user/:id** - Get threads by user ID
- **PUT /threads/:id** - Update thread
- **DELETE /threads/:id** - Delete thread
- **POST /threads/:id/like** - Like thread
- **DELETE /threads/:id/like** - Unlike thread
- **POST /threads/:id/bookmark** - Bookmark thread
- **DELETE /threads/:id/bookmark** - Remove bookmark
- **POST /threads/:id/repost** - Repost thread
- **DELETE /threads/:id/repost** - Remove repost
- **GET /categories** - Get all categories
- **POST /categories** - Create new category
- **GET /categories/:id** - Get category by ID
- **PUT /categories/:id** - Update category
- **DELETE /categories/:id** - Delete category
- **POST /threads/:id/categories/:categoryId** - Add category to thread
- **DELETE /threads/:id/categories/:categoryId** - Remove category from thread
- **GET /threads/:id/categories** - Get thread categories

### 3. Community Service

Manages communities, membership, and moderation.

#### Interfaces

```go
// CommunityService defines operations for community management
type CommunityService interface {
    CreateCommunity(ctx context.Context, name, description string, userID string, isPrivate bool) (*model.Community, error)
    GetCommunityByID(ctx context.Context, communityID string) (*model.Community, error)
    GetAllCommunities(ctx context.Context, page, limit int) ([]*model.Community, error)
    UpdateCommunity(ctx context.Context, communityID, userID string, updates map[string]interface{}) (*model.Community, error)
    DeleteCommunity(ctx context.Context, communityID, userID string) error
    JoinCommunity(ctx context.Context, communityID, userID string) error
    LeaveCommunity(ctx context.Context, communityID, userID string) error
    IsMember(ctx context.Context, communityID, userID string) (bool, error)
    GetCommunityMembers(ctx context.Context, communityID string, page, limit int) ([]*model.User, error)
    GetUserCommunities(ctx context.Context, userID string, page, limit int) ([]*model.Community, error)
}
```

#### Endpoints

- **POST /communities** - Create new community
- **GET /communities/:id** - Get community by ID
- **GET /communities** - Get all communities with pagination
- **PUT /communities/:id** - Update community
- **DELETE /communities/:id** - Delete community
- **POST /communities/:id/join** - Join community
- **DELETE /communities/:id/leave** - Leave community
- **GET /communities/:id/members** - Get community members
- **GET /users/:id/communities** - Get user's communities

### 4. AI Service

Provides AI capabilities for content analysis and recommendations.

#### Endpoints

- **GET /health** - Health check for service
- **GET /categories** - Get all supported categories
- **POST /predict/category** - Predict category based on content

## Data Models

### Thread Model

```go
type Thread struct {
    ThreadID        uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
    UserID          uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
    Content         string     `gorm:"type:text;not null"`
    IsPinned        bool       `gorm:"default:false"`
    WhoCanReply     string     `gorm:"type:varchar(20);not null"`
    ScheduledAt     *time.Time `gorm:"type:timestamp with time zone"`
    CommunityID     *uuid.UUID `gorm:"type:uuid"`
    IsAdvertisement bool       `gorm:"default:false"`
    CreatedAt       time.Time  `gorm:"autoCreateTime"`
    UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
    DeletedAt       *time.Time `gorm:"index"`
}
```

### Category Model

```go
type Category struct {
    CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
    Name       string     `gorm:"type:varchar(50);not null"`
    Type       string     `gorm:"type:varchar(10);not null"`
    CreatedAt  time.Time  `gorm:"autoCreateTime"`
    UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
    DeletedAt  *time.Time `gorm:"index"`
}

type ThreadCategory struct {
    ThreadID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
    CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
    DeletedAt  *time.Time `gorm:"index"`
}
```

## Automatic Thread Categorization

The system uses a machine learning model to automatically suggest categories for threads based on their content. The AI service analyzes the text and provides category predictions with confidence scores.

Category types include:
- Technology
- Health
- Education
- Entertainment 
- Science
- Sports
- Politics
- Business
- Lifestyle
- Travel
- Other

The category suggestion is integrated in the thread creation flow, with higher confidence predictions being auto-selected.

## API Gateway

The API gateway routes requests to the appropriate microservices and handles cross-cutting concerns:

- Authentication and authorization
- Request logging
- CORS
- Rate limiting
- Request validation

## Event Bus

The event bus enables asynchronous communication between services using RabbitMQ:

- Thread creation/deletion events
- User action events (likes, reposts, etc.)
- Community events
- Notification events 