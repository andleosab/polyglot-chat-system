-- 1. Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Custom Type
CREATE TYPE conversation_type AS ENUM ('private', 'group');

-- 2. conversations Table
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NULL,
    type conversation_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    last_message_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_conversation_name CHECK (
        (type = 'private' AND name IS NULL) OR
        (type = 'group' AND name IS NOT NULL)
    )    
);

-- 3. participants Table (SIMPLIFIED)
CREATE TABLE participants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    
    -- FK to conversations table (BIGINT for speed)
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    -- Public UUID of the user (No separate participant_uuid needed)
    user_uuid UUID NOT NULL, 
    username VARCHAR(100) NOT NULL,
    
    is_admin BOOLEAN DEFAULT FALSE NOT NULL, 
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Enforces that a user can only be in a conversation once
    UNIQUE (conversation_id, user_uuid) 
);
CREATE INDEX idx_participants_conversation_id ON participants(conversation_id);
CREATE INDEX idx_participants_user_uuid ON participants(user_uuid);

-- 4. messages Table
CREATE TABLE messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    
    -- FK to conversations table
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    sender_user_uuid UUID NOT NULL, 
    
    content TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    client_ts BIGINT
);

CREATE INDEX idx_messages_conversation_time ON messages (conversation_id, sent_at DESC);

-- for Lateral Join optimization in GetUserConversations query
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, id DESC);

-- For filtering conversations by type
CREATE INDEX idx_conversations_type ON conversations(type);

-- Composite index — the most important one for both queries
CREATE INDEX idx_participants_conv_user ON participants(conversation_id, user_uuid);

