-- Fix Chat Schema Part 2 (Revised)
-- This script fixes the remaining issues with the chat database schema

-- Fix primary key names
BEGIN;
ALTER TABLE chat_participants 
    DROP CONSTRAINT IF EXISTS chat_participants_new_pkey,
    ADD CONSTRAINT chat_participants_pkey PRIMARY KEY (chat_id, user_id);
COMMIT;

-- Fix deleted_chats primary key
BEGIN;
ALTER TABLE deleted_chats 
    DROP CONSTRAINT IF EXISTS deleted_chats_new_pkey,
    ADD CONSTRAINT deleted_chats_pkey PRIMARY KEY (chat_id, user_id);
COMMIT;

-- Add missing foreign key constraint to deleted_chats
BEGIN;
ALTER TABLE deleted_chats
    ADD CONSTRAINT deleted_chats_chat_id_fkey 
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE;
COMMIT;

-- Create backup of messages table
BEGIN;
CREATE TABLE IF NOT EXISTS messages_backup AS SELECT * FROM messages;
COMMIT;

-- Update messages table structure
BEGIN;
ALTER TABLE messages
    ALTER COLUMN sent_at SET DEFAULT NOW();
COMMIT;

BEGIN;
ALTER TABLE messages
    ALTER COLUMN sent_at SET NOT NULL;
COMMIT;

BEGIN;
ALTER TABLE messages
    ALTER COLUMN is_read SET DEFAULT FALSE;
COMMIT;

BEGIN;
ALTER TABLE messages
    ALTER COLUMN is_edited SET DEFAULT FALSE;
COMMIT;

BEGIN;
ALTER TABLE messages
    ALTER COLUMN is_deleted SET DEFAULT FALSE;
COMMIT;

-- Add foreign key to chat_id
BEGIN;
ALTER TABLE messages
    ADD CONSTRAINT messages_chat_id_fkey
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE;
COMMIT;

-- Note: Skip the users foreign key constraint as the users table might be in another database schema

-- Add text search index
BEGIN;
DROP INDEX IF EXISTS idx_messages_content;
CREATE INDEX idx_messages_content ON messages USING gin (to_tsvector('english', content));
COMMIT;

-- Add sent_at index
BEGIN;
DROP INDEX IF EXISTS idx_messages_sent_at;
CREATE INDEX idx_messages_sent_at ON messages(sent_at);
COMMIT; 