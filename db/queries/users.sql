-- name: InsertUser :one
INSERT INTO users (username, email, password,password_salt)
VALUES (@username, @email, @password, @password_salt)
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