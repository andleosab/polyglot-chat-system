ALTER TABLE conversations 
ADD CONSTRAINT chk_conversation_name 
CHECK (
    (type = 'private' AND name IS NULL) OR 
    (type = 'group' AND name IS NOT NULL)
);