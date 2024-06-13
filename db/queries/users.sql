-- name: InsertUser :one
INSERT INTO users (username, email, password,password_salt)
VALUES (:username, :email, :password, :password_salt)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = ?
LIMIT 1;