-- name: TrackRedirect :exec
INSERT INTO visits (url_id, ip_address, user_agent)
VALUES (?, ?, ?);

-- name: ListStatistics :many
SELECT
	count(v.id) as visits,
	CAST(MIN(u.id) as INTEGER) as id,
	u.short_url as short_url,
	CAST(MIN(u.long_url) as TEXT) as long_url
FROM
	urls u
LEFT JOIN visits v ON u.id = v.url_id
WHERE
	u.author_id = ?
GROUP BY
	u.short_url
ORDER BY
	visits