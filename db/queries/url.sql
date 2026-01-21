-- name: AddShortURL :one
INSERT INTO urls (title,short_url, long_url, author_id) 
VALUES (@title, @short_url, @long_url, @author_id)
RETURNING *;

-- name: ListShortURLs :many
SELECT *
FROM urls
WHERE urls.author_id = @author_id
AND urls.is_archived = @is_archived
ORDER BY id DESC;

-- name: GetShortURL :one
SELECT *
FROM urls
WHERE urls.short_url = @short_url;

-- name: GetByID :one
SELECT *
FROM urls
WHERE urls.id = @id;

-- name: DeleteURL :exec
DELETE FROM urls
WHERE id = @id
AND urls.author_id = @author_id;

-- name: UpdateTitle :one
UPDATE urls
SET title = @title
WHERE urls.short_url = @short_url
AND urls.author_id = @author_id
RETURNING *;

-- name: ArchiveURL :exec
UPDATE urls
SET is_archived = true
WHERE urls.short_url = @short_url
AND urls.author_id = @author_id;

-- name: UnarchiveURL :exec
UPDATE urls
SET is_archived = false
WHERE urls.short_url = @short_url
AND urls.author_id = @author_id;

-- name: UpdateURLStatus :exec
UPDATE urls
SET is_active = @is_active
WHERE urls.short_url = @short_url;