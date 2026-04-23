-- name: AddParticipant :one
INSERT INTO participants (conversation_id, user_uuid, username, is_admin) 
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetConversationMembers :many
SELECT id, conversation_id, user_uuid, username, is_admin, joined_at FROM participants WHERE conversation_id = $1;

/*-- name: GetUserConversations :many
SELECT t1.id, t1.name, t1.type, t1.last_message_at 
FROM conversations t1 
JOIN participants t2 ON t1.id = t2.conversation_id 
WHERE t2.user_uuid = $1 
ORDER BY t1.last_message_at DESC;
*/

-- name: UpdateAdminRole :exec
UPDATE participants SET is_admin = $3 WHERE conversation_id = $1 AND user_uuid = $2;

-- name: RemoveParticipant :exec
DELETE FROM participants WHERE conversation_id = $1 AND user_uuid = $2;