-- name: InsertModerationFlag :one
INSERT INTO moderation_flags (url_id, user_id, risk_score, threat_type, status)
VALUES (@url_id, @user_id, @risk_score, @threat_type, 'flagged')
RETURNING *;

-- name: ListModerationFlags :many
SELECT 
    f.id,
    f.url_id,
    f.user_id,
    f.risk_score,
    f.threat_type,
    f.status,
    f.created_at,
    f.reviewed_at,
    f.reviewed_by,
    urls.long_url,
    urls.short_url,
    users.username,
    users.email
FROM moderation_flags f
JOIN urls ON f.url_id = urls.id
JOIN users ON f.user_id = users.id
ORDER BY f.created_at DESC;

-- name: UpdateModerationFlagStatus :exec
UPDATE moderation_flags
SET status = @status,
	reviewed_at = CURRENT_TIMESTAMP,
	reviewed_by = @reviewed_by
WHERE id = @id;

-- name: GetModerationFlagByID :one
SELECT *
FROM moderation_flags
WHERE id = @id;
