-- name: InsertUserOauth :one
INSERT INTO users (guid, username, email, is_oauth)
VALUES (@guid, @username, @email, true)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = @email
LIMIT 1;

-- name: InsertOauth2State :exec
INSERT INTO oauth2_state (state, provider)
VALUES (@state, @provider);

-- name: GetOauth2State :one
SELECT *
FROM oauth2_state
WHERE state = @state
AND expire_at > CURRENT_TIMESTAMP
LIMIT 1;