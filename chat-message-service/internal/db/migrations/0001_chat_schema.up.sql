-- 1. Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Custom Type
CREATE TYPE conversation_type AS ENUM ('private', 'group');

-- 2. conversations Table
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type conversation_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMPTZ
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
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE INDEX idx_messages_conversation_time ON messages (conversation_id, sent_at DESC);



