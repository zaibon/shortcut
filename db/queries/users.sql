-- name: InsertUser :one
INSERT INTO users (username, email, password, password_salt)
VALUES (@username, @email, @password, @password_salt)
RETURNING *;

-- name: InsertUserOauth :one
INSERT INTO users (username, email, is_oauth)
VALUES (@username, @email, true)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = @email
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = @username, email = @email
WHERE id = @id
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET password = @password, password_salt = @password_salt
WHERE id = @id;

-- name: InsertOauth2State :exec
INSERT INTO oauth2_state (state)
VALUES (@state);

-- name: GetOauth2State :one
SELECT *
FROM oauth2_state
WHERE state = @state
AND expire_at > CURRENT_TIMESTAMP
LIMIT 1;