-- name: CreateConversation :one
INSERT INTO conversations (type, name, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConversationByID :one
SELECT * FROM conversations
WHERE id = $1;

-- name: UpdateConversationName :exec
UPDATE conversations
SET name = $1
WHERE id = $2;

-- name: DeleteConversation :exec
DELETE FROM conversations
WHERE id = $1;

-- name: UpdateConversationLastMessageAt :exec
UPDATE conversations SET last_message_at = NOW() WHERE id = $1;

-- name: GetUserConversations :many
SELECT 
    c.id,
    c.type,
    c.created_at,
    c.created_by,
    COALESCE(c.name, (
        SELECT p2.username 
        FROM participants p2 
        WHERE p2.conversation_id = c.id AND p2.user_uuid != $1
        limit 1
    )) AS display_name,
    COALESCE(m.content, '') AS last_message,
    COALESCE(m.sent_at, '1970-01-01 00:00:00+00'::timestamptz) AS sent_at    
    --m.content AS last_message,
    --m.sent_at
FROM conversations c
JOIN participants p ON c.id = p.conversation_id
LEFT JOIN LATERAL (
    SELECT m2.content, m2.sent_at
    FROM messages m2
    WHERE m2.conversation_id = c.id
    ORDER BY m2.id DESC
    LIMIT 1
) m ON true
WHERE p.user_uuid  = $1
ORDER BY m.sent_at DESC;

-- name: GetUserConversationIDs :many
SELECT 
    c.id
FROM conversations c
JOIN participants p ON c.id = p.conversation_id
WHERE p.user_uuid = $1;

-- name: GetUserGroups :many
SELECT 
    c.id,
    c.name AS group_name,
    c.created_at
FROM conversations c
JOIN participants p 
    ON p.conversation_id = c.id
WHERE p.user_uuid = $1
  AND c.type = 'group'
ORDER BY c.name;

-- name: GetPrivateConversationID :one
SELECT p.conversation_id
FROM participants p
JOIN conversations c ON c.id = p.conversation_id
WHERE c.type = 'private'
AND p.user_uuid IN (@user_uuid1, @user_uuid2)
GROUP BY p.conversation_id
HAVING COUNT(*) = 2
LIMIT 1;
