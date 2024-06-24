-- name: InsertUser :one
INSERT INTO users (guid, username, email, password, password_salt)
VALUES (@guid, @username, @email, @password, @password_salt)
RETURNING *;

-- name: InsertUserOauth :one
INSERT INTO users (guid, username, email, is_oauth)
VALUES (@guid, @username, @email, true)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = @email
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = @username, email = @email
WHERE guid = @guid
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET password = @password, password_salt = @password_salt
WHERE guid = @guid;

-- name: InsertOauth2State :exec
INSERT INTO oauth2_state (state)
VALUES (@state);

-- name: GetOauth2State :one
SELECT *
FROM oauth2_state
WHERE state = @state
AND expire_at > CURRENT_TIMESTAMP
LIMIT 1;