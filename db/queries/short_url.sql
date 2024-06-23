-- name: AddShortURL :one
INSERT INTO urls (title,short_url, long_url, author_id) 
VALUES (@title, @short_url, @long_url, @author_id)
RETURNING *;

-- name: ListShortURLs :many
SELECT *
FROM urls
WHERE urls.author_id = @author_id
ORDER BY 
    CASE WHEN @sort_by = 'title_asc' THEN title END ASC,
    CASE WHEN @sort_by = 'title_desc' THEN title END DESC,
    CASE WHEN @sort_by = 'created_at_asc' THEN created_at END ASC,
    CASE WHEN @sort_by = 'created_at_desc' THEN created_at END DESC
;

-- name: GetShortURL :one
SELECT *
FROM urls
WHERE urls.short_url = @short_url;