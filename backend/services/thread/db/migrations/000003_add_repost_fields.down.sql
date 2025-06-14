-- Remove indexes
DROP INDEX IF EXISTS idx_reposts_new_thread_id;
DROP INDEX IF EXISTS idx_threads_original_thread_id;

-- Remove foreign key constraints
ALTER TABLE reposts
DROP CONSTRAINT IF EXISTS fk_reposts_new_thread;

ALTER TABLE threads
DROP CONSTRAINT IF EXISTS fk_threads_original_thread;

-- Remove columns from reposts table
ALTER TABLE reposts
DROP COLUMN IF EXISTS new_thread_id;

-- Remove columns from threads table
ALTER TABLE threads
DROP COLUMN IF EXISTS original_thread_id,
DROP COLUMN IF EXISTS is_repost; 