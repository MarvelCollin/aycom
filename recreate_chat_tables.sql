-- Recreate Chat Tables from Schema
-- This script recreates the chat tables based on the proper schema

-- Drop existing tables in the right order to avoid foreign key issues
DROP TABLE IF EXISTS deleted_chats CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS chat_participants CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS test_chat CASCADE; 
DROP TABLE IF EXISTS test_participant CASCADE;

-- Chat conversations (individual or group)
CREATE TABLE chats (
    chat_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(100), -- For group chats
    created_by UUID NOT NULL, -- User who created the chat
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE -- For soft delete
);

-- Chat participants (users in a chat)
CREATE TABLE chat_participants (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

-- Chat messages
CREATE TABLE messages (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT,
    media_url VARCHAR(512),
    media_type VARCHAR(10) CHECK (media_type IN ('Image', 'GIF', 'Video')),
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    unsent BOOLEAN DEFAULT FALSE NOT NULL,
    unsent_at TIMESTAMP WITH TIME ZONE,
    deleted_for_sender BOOLEAN DEFAULT FALSE NOT NULL,
    deleted_for_all BOOLEAN DEFAULT FALSE NOT NULL,
    reply_to_message_id UUID,
    is_read BOOLEAN DEFAULT FALSE,
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE, -- For soft delete
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
    -- Note: sender_id foreign key to users table omitted as users may be in another database
);

-- Index for searching messages
CREATE INDEX idx_messages_content ON messages USING gin (to_tsvector('english', content));

-- Track deleted conversations for users (soft delete)
CREATE TABLE deleted_chats (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

-- Track message unsend window (for enforcing 1 minute unsend)
CREATE INDEX idx_messages_sent_at ON messages(sent_at);

-- Additional indexes for performance
CREATE INDEX idx_chat_participants_user_id ON chat_participants(user_id);
CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_deleted_chats_user_id ON deleted_chats(user_id);

-- Create a test chat with its participant to verify everything works
INSERT INTO chats (chat_id, is_group, name, created_by)
VALUES (gen_random_uuid(), false, 'Test Chat', '00000000-0000-0000-0000-000000000001');

-- Get the chat_id we just created
DO $$
DECLARE
    v_chat_id UUID;
BEGIN
    SELECT chat_id INTO v_chat_id FROM chats LIMIT 1;
    
    -- Add a test participant to the chat
    INSERT INTO chat_participants (chat_id, user_id, is_admin)
    VALUES (v_chat_id, '00000000-0000-0000-0000-000000000001', true);
END $$; 