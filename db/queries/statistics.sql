-- name: TrackRedirect :one
INSERT INTO visits (url_id, ip_address, user_agent)
VALUES (?, ?, ?)
RETURNING *;

-- name: InsertVisitLocation :one
INSERT INTO visit_locations (
	visit_id,
	address,
	country_code,
	country_name,
	subdivision,
	continent,
	city_name,
	latitude,
	longitude,
	source
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListStatisticsPerAuthor :many
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
	visits;

-- name: ListVisits :many
SELECT *
FROM visits
ORDER BY id DESC;