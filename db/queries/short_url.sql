-- name: AddShortURL :one
INSERT INTO urls (short_url, long_url, author_id) 
VALUES (?, ?, ?)
RETURNING *;

-- name: ListShortURLs :many
SELECT *
FROM urls
WHERE urls.author_id = ?
ORDER BY id DESC;

-- name: GetShortURL :one
SELECT *
FROM urls
WHERE urls.short_url = ?;