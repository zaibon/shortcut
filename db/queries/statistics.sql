-- name: TrackRedirect :one
INSERT INTO visits (url_id, ip_address, user_agent)
VALUES (@url_id, @ip_address, @user_agent)
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
VALUES (
	@visit_id,
	@address,
	@country_code,
	@country_name,
	@subdivision,
	@continent,
	@city_name,
	@latitude,
	@longitude,
	@source
)
RETURNING *;

-- name: ListStatisticsPerAuthor :many
SELECT
	count(v.id) as nr_visits,
	MIN(u.id)::INTEGER as id,
	u.short_url as short_url,
	MIN(u.title):: TEXT as title,
	MIN(u.long_url):: TEXT as long_url,
	MIN(u.created_at)::TIMESTAMP as created_at
FROM
	urls u
LEFT JOIN visits v ON u.id = v.url_id
WHERE
	u.author_id = @author_id
GROUP BY
	u.short_url, u.id
ORDER BY
	u.id DESC;

-- name: StatisticPerURL :one
SELECT
	count(v.id) as nr_visits,
	MIN(u.id)::INTEGER as id,
	u.short_url as short_url,
	MIN(u.title):: TEXT as title,
	MIN(u.long_url):: TEXT as long_url,
	MIN(u.created_at)::TIMESTAMP as created_at
FROM
	urls u
LEFT JOIN visits v ON u.id = v.url_id
WHERE
	u.short_url = @short_url
	AND u.author_id = @author_id
GROUP BY
	u.short_url, u.id
LIMIT 1;

-- name: ListVisits :many
SELECT *
FROM visits
ORDER BY id DESC;