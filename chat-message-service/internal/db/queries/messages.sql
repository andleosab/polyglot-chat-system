-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_user_uuid, content, client_ts) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateMessageContent :exec
UPDATE messages SET content = $1, edited_at = CURRENT_TIMESTAMP WHERE id = $2;

-- name: GetMessagesByConversation :many
SELECT 
    m.id,
    m.conversation_id,
    m.content, 
    m.sent_at, 
    m.edited_at,
    m.is_deleted,
    s.username AS sender_username,
    m.sender_user_uuid,
    m.client_ts
FROM messages m 
JOIN participants s ON m.conversation_id = s.conversation_id AND m.sender_user_uuid = s.user_uuid 
JOIN participants v ON v.conversation_id = m.conversation_id AND v.user_uuid = $4
WHERE m.conversation_id = $1 AND m.is_deleted = FALSE
AND m.sent_at > v.joined_at 
ORDER BY m.sent_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetMessagesByConversationCursor :many
SELECT 
    m.id,
    m.sent_at,
    s.joined_at,
    m.conversation_id,
    m.content, 
    m.sent_at, 
    m.edited_at,
    m.is_deleted,
    s.username AS sender_username,
    m.sender_user_uuid,
    m.client_ts
FROM messages m 
JOIN participants s ON m.conversation_id = s.conversation_id AND m.sender_user_uuid = s.user_uuid
JOIN participants v ON v.conversation_id = m.conversation_id AND v.user_uuid = @user_uuid
WHERE m.conversation_id = @conversation_id AND m.is_deleted = FALSE
AND m.id < @message_id
AND m.sent_at > v.joined_at 
ORDER BY m.sent_at DESC
LIMIT $1;

-- name: DeleteMessage :exec
UPDATE messages SET is_deleted = TRUE WHERE id = $1;