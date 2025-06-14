-- Add repost fields to the threads table
ALTER TABLE threads
ADD COLUMN is_repost BOOLEAN DEFAULT false,
ADD COLUMN original_thread_id UUID NULL;

-- Add foreign key constraint
ALTER TABLE threads
ADD CONSTRAINT fk_threads_original_thread
FOREIGN KEY (original_thread_id) REFERENCES threads(thread_id);

-- Add new_thread_id column to reposts table
ALTER TABLE reposts
ADD COLUMN new_thread_id UUID NULL;

-- Add foreign key constraint
ALTER TABLE reposts
ADD CONSTRAINT fk_reposts_new_thread
FOREIGN KEY (new_thread_id) REFERENCES threads(thread_id);

-- Add index for original_thread_id
CREATE INDEX idx_threads_original_thread_id ON threads(original_thread_id);

-- Add index for new_thread_id in reposts
CREATE INDEX idx_reposts_new_thread_id ON reposts(new_thread_id); 