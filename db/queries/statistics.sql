-- name: TrackRedirect :exec
INSERT INTO visits (url_id, ip_address, user_agent)
VALUES (?, ?, ?);