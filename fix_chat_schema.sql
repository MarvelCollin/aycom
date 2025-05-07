-- Fix Chat Schema Migration
-- This script corrects the discrepancies between the intended schema and the actual database

BEGIN;

-- Fix chat_participants table
DROP TABLE IF EXISTS chat_participants_new;
CREATE TABLE chat_participants_new (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

-- Migrate data from old table (converting text to UUID)
INSERT INTO chat_participants_new (chat_id, user_id, joined_at, is_admin)
SELECT 
    CAST(chat_id AS UUID), 
    CAST(user_id AS UUID), 
    COALESCE(joined_at, NOW()), 
    FALSE
FROM chat_participants
WHERE 
    chat_id IS NOT NULL 
    AND user_id IS NOT NULL 
    AND chat_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
    AND user_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Fix deleted_chats table
DROP TABLE IF EXISTS deleted_chats_new;
CREATE TABLE deleted_chats_new (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

-- Migrate data from old table (converting text to UUID)
INSERT INTO deleted_chats_new (chat_id, user_id, deleted_at)
SELECT 
    CAST(chat_id AS UUID), 
    CAST(user_id AS UUID), 
    COALESCE(deleted_at, NOW())
FROM deleted_chats
WHERE 
    chat_id IS NOT NULL 
    AND user_id IS NOT NULL 
    AND chat_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
    AND user_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Drop old tables and rename new ones
DROP TABLE chat_participants;
ALTER TABLE chat_participants_new RENAME TO chat_participants;

DROP TABLE deleted_chats;
ALTER TABLE deleted_chats_new RENAME TO deleted_chats;

-- Create necessary indexes
CREATE INDEX idx_chat_participants_user_id ON chat_participants(user_id);
CREATE INDEX idx_deleted_chats_user_id ON deleted_chats(user_id);

-- Remove test tables if needed
-- DROP TABLE IF EXISTS test_chat;
-- DROP TABLE IF EXISTS test_participant;

-- Verify structure is correct
ANALYZE chat_participants;
ANALYZE deleted_chats;

COMMIT; 