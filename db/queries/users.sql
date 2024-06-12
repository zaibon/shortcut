-- name: InsertUser :one
INSERT INTO users (username, email, password)
VALUES (:username, :email, :password)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = ?
LIMIT 1;