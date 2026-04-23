-- for Lateral Join optimization in GetUserConversations query
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, id DESC);

-- For filtering conversations by type
CREATE INDEX idx_conversations_type ON conversations(type);

-- Composite index — the most important one for both queries
CREATE INDEX idx_participants_conv_user ON participants(conversation_id, user_uuid);