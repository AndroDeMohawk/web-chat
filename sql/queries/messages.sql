-- name: CreateMessage :one
INSERT INTO messages (
    user_id, content
) VALUES (
             $1, $2
         )
RETURNING id, user_id, content, created_at;

-- name: GetRecentMessages :many
SELECT
    m.id,
    m.content,
    m.created_at,
    u.id AS user_id,
    u.username
FROM messages m
         JOIN users u ON m.user_id = u.id
ORDER BY m.created_at DESC
LIMIT $1;

-- name: ClearAllMessages :exec
DELETE FROM messages;