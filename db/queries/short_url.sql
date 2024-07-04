-- name: AddShortURL :one
INSERT INTO urls (title,short_url, long_url, author_id) 
VALUES (@title, @short_url, @long_url, @author_id)
RETURNING *;

-- name: ListShortURLs :many
SELECT *
FROM urls
WHERE urls.author_id = @author_id
ORDER BY id DESC;

-- name: GetShortURL :one
SELECT *
FROM urls
WHERE urls.short_url = @short_url;

-- name: UpdateTitle :exec
UPDATE urls
SET title = @title
WHERE urls.short_url = @short_url
AND urls.author_id = @author_id;