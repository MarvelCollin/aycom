-- USERS SERVICE

-- Security questions for account recovery
CREATE TABLE security_questions (
    question_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_text VARCHAR(255) NOT NULL UNIQUE
);
-- Relation: One-to-Many with users (One question can be used by many users)

-- Core user data
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    password_salt VARCHAR(64) NOT NULL,
    profile_picture_url VARCHAR(512),
    banner_url VARCHAR(512),
    bio TEXT,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female')),
    date_of_birth DATE NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    is_activated BOOLEAN DEFAULT FALSE NOT NULL,
    is_banned BOOLEAN DEFAULT FALSE NOT NULL,
    is_deactivated BOOLEAN DEFAULT FALSE NOT NULL,
    is_private BOOLEAN DEFAULT FALSE NOT NULL,
    is_premium BOOLEAN DEFAULT FALSE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE NOT NULL,
    newsletter_subscription BOOLEAN DEFAULT FALSE NOT NULL,
    security_question_id UUID NOT NULL,
    security_answer VARCHAR(255) NOT NULL,
    google_id VARCHAR(100),
    last_login_at TIMESTAMP WITH TIME ZONE,
    refresh_token VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (security_question_id) REFERENCES security_questions(question_id)
);
-- Relations:
-- 1. Many-to-One with security_questions (Many users use one security question)
-- 2. One-to-Many with threads, replies, likes, etc. (One user can create many content items)
-- 3. One-to-One with user_settings (One user has one settings record)

-- Represents follow relationships between users
CREATE TABLE followers (
    follower_id UUID NOT NULL,  -- User who follows someone
    followee_id UUID NOT NULL,  -- User who is being followed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (follower_id, followee_id), 
    FOREIGN KEY (follower_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CHECK (follower_id != followee_id)
);
-- Relation: Many-to-Many between users (Users can follow many users and be followed by many users)

-- Represents block relationships between users
CREATE TABLE blocked_users (
    blocker_id UUID NOT NULL,  -- User who blocks someone
    blocked_id UUID NOT NULL,  -- User who is blocked
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (blocker_id, blocked_id),
    FOREIGN KEY (blocker_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CHECK (blocker_id != blocked_id)
);
-- Relation: Many-to-Many between users (Users can block many users and be blocked by many users)

-- User preference settings
CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY,
    font_size VARCHAR(20) DEFAULT 'medium' NOT NULL,
    font_color VARCHAR(20) DEFAULT 'default' NOT NULL,
    notification_like BOOLEAN DEFAULT TRUE NOT NULL,
    notification_repost BOOLEAN DEFAULT TRUE NOT NULL,
    notification_follow BOOLEAN DEFAULT TRUE NOT NULL,
    notification_mention BOOLEAN DEFAULT TRUE NOT NULL,
    notification_community BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
-- Relation: One-to-One with users (One user has one settings record)

-- COMMUNITIES SERVICE

-- Community information
CREATE TABLE communities (
    community_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    logo_url VARCHAR(512) NOT NULL,
    banner_url VARCHAR(512) NOT NULL,
    creator_id UUID NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (creator_id) REFERENCES users(user_id) ON DELETE CASCADE
);
-- Relations:
-- 1. Many-to-One with users (Many communities can be created by one user)
-- 2. One-to-Many with threads (One community can have many threads)
-- 3. One-to-Many with community_members (One community can have many members)

-- CONTENT SERVICE

-- Thread posts
CREATE TABLE threads (
    thread_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE NOT NULL,
    who_can_reply VARCHAR(20) NOT NULL CHECK (who_can_reply IN ('Everyone', 'Accounts You Follow', 'Verified Accounts')),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    community_id UUID,
    is_advertisement BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE SET NULL
);
-- Relations:
-- 1. Many-to-One with users (Many threads can be created by one user)
-- 2. Many-to-One with communities (Many threads can belong to one community)
-- 3. One-to-Many with replies (One thread can have many replies)
-- 4. One-to-One with polls (One thread can have one poll)

-- Thread replies
CREATE TABLE replies (
    reply_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE NOT NULL,
    parent_reply_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE
);
-- Relations:
-- 1. Many-to-One with threads (Many replies can belong to one thread)
-- 2. Many-to-One with users (Many replies can be created by one user)
-- 3. Many-to-One with replies (Many replies can be responses to one parent reply - nested replies)

-- Media attachments for threads and replies
CREATE TABLE media (
    media_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID,
    reply_id UUID,
    type VARCHAR(10) NOT NULL CHECK (type IN ('Image', 'GIF', 'Video')),
    url VARCHAR(512) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE,
    CHECK ((thread_id IS NULL AND reply_id IS NOT NULL) OR (thread_id IS NOT NULL AND reply_id IS NULL))
);
-- Relations:
-- 1. Many-to-One with threads (Many media items can belong to one thread)
-- 2. Many-to-One with replies (Many media items can belong to one reply)

-- Categories for threads and communities
CREATE TABLE categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('Thread', 'Community')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE (name, type)
);
-- Used for categorizing both threads and communities

-- Junction table to connect threads with categories
CREATE TABLE thread_categories (
    thread_id UUID NOT NULL,
    category_id UUID NOT NULL,
    PRIMARY KEY (thread_id, category_id),
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between threads and categories

-- Junction table to connect communities with categories
CREATE TABLE community_categories (
    community_id UUID NOT NULL,
    category_id UUID NOT NULL,
    PRIMARY KEY (community_id, category_id),
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between communities and categories

-- Represents like actions on threads or replies
CREATE TABLE likes (
    user_id UUID NOT NULL,
    thread_id UUID,
    reply_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (user_id, COALESCE(thread_id, uuid_nil()), COALESCE(reply_id, uuid_nil())),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE,
    CHECK ((thread_id IS NULL AND reply_id IS NOT NULL) OR (thread_id IS NOT NULL AND reply_id IS NULL))
);
-- Relations:
-- 1. Many-to-Many between users and threads (Users can like many threads, threads can be liked by many users)
-- 2. Many-to-Many between users and replies (Users can like many replies, replies can be liked by many users)

-- Represents repost actions
CREATE TABLE reposts (
    user_id UUID NOT NULL,
    thread_id UUID NOT NULL,
    repost_text TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (user_id, thread_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between users and threads (Users can repost many threads, threads can be reposted by many users)

-- Represents bookmark actions
CREATE TABLE bookmarks (
    user_id UUID NOT NULL,
    thread_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (user_id, thread_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between users and threads (Users can bookmark many threads, threads can be bookmarked by many users)

-- Hashtags used in threads
CREATE TABLE hashtags (
    hashtag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
-- Will be connected to threads via the thread_hashtags junction table

-- Junction table to connect threads with hashtags
CREATE TABLE thread_hashtags (
    thread_id UUID NOT NULL,
    hashtag_id UUID NOT NULL,
    PRIMARY KEY (thread_id, hashtag_id),
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (hashtag_id) REFERENCES hashtags(hashtag_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between threads and hashtags

-- Represents user mentions in threads or replies
CREATE TABLE user_mentions (
    mention_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mentioned_user_id UUID NOT NULL,  -- User who is mentioned
    thread_id UUID,                  -- Thread where the mention occurs (if in a thread)
    reply_id UUID,                   -- Reply where the mention occurs (if in a reply)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (mentioned_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE,
    CHECK ((thread_id IS NULL AND reply_id IS NOT NULL) OR (thread_id IS NOT NULL AND reply_id IS NULL))
);
-- Relations:
-- 1. Many-to-One with users (Many mentions can reference one user)
-- 2. Many-to-One with threads (Many mentions can be in one thread)
-- 3. Many-to-One with replies (Many mentions can be in one reply)

-- COMMUNITY MANAGEMENT SERVICE

-- Users who are members of communities
CREATE TABLE community_members (
    community_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(10) NOT NULL CHECK (role IN ('Owner', 'Moderator', 'Member')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (community_id, user_id),
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
-- Relation: Many-to-Many between communities and users with roles

-- Requests to join communities
CREATE TABLE community_join_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('Pending', 'Approved', 'Rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE (community_id, user_id)
);
-- Relations:
-- 1. Many-to-One with communities (Many requests can be for one community)
-- 2. Many-to-One with users (Many requests can be from one user)

-- Rules for communities
CREATE TABLE community_rules (
    rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL,
    rule_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE
);
-- Relation: Many-to-One with communities (Many rules can belong to one community)

--