-- name: TrackRedirect :one
INSERT INTO visits (url_id, ip_address, user_agent, browser_id, referrer)
VALUES (@url_id, @ip_address, @user_agent, @browser_id, @referrer)
RETURNING *;


-- name: UpsertBrowser :one
INSERT INTO browsers (name, version, platform, mobile)
VALUES (@name, @version, @platform, @mobile)
ON CONFLICT (name, version, platform, mobile) 
DO UPDATE SET name = EXCLUDED.name
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
	AND (
		@search = ''
		-- OR to_tsvector(regexp_replace(long_url, '^(https?://)?(www\.)?', '', 'i')) @@ to_tsquery(@search)
		OR long_url ILIKE '%' || @search || '%'
	)
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

-- name: CountURLThisMonth :one
SELECT count(*)
FROM urls
WHERE author_id = @author_id
AND created_at BETWEEN date_trunc('month', CURRENT_DATE) AND date_trunc('month', CURRENT_DATE) + INTERVAL '1 month';


-- name: CountTotalVisitThisMonth :one
SELECT count(visits.*)
FROM visits
JOIN urls ON urls.id = visits.url_id
WHERE urls.author_id = @author_id
AND created_at BETWEEN date_trunc('month', CURRENT_DATE) AND date_trunc('month', CURRENT_DATE) + INTERVAL '1 month';

-- SQL query to get the location distribution data for a specific URL
-- name: LocationDistribution :many
SELECT
    vl.country_code,
    vl.country_name,
    count(vl.visit_id) AS visit_count,
    (count(vl.visit_id) * 100.0 / total_visits.total)::float AS percentage
FROM
    visit_locations vl
JOIN
    visits v ON vl.visit_id = v.id
JOIN
    urls u ON v.url_id = u.id
CROSS JOIN (SELECT count(*) as total FROM visits WHERE visits.url_id = @url_id) AS total_visits
WHERE
	u.author_id = @author_id
AND
    u.id = @url_id -- Replace 'your_short_url' with the actual short URL
GROUP BY
    vl.country_code, vl.country_name, total_visits.total
ORDER BY
    visit_count DESC;


-- SQL query to get the distribution of browsers for a specific URL
-- name: BrowserDistribution :many
SELECT
	browsers.name,
	browsers.version,
	browsers.platform,
  	browsers.mobile,
    count(v.id) AS visit_count,
    (count(v.id) * 100.0 / total_visits.total)::float AS percentage
FROM
    visits v
JOIN
    urls u ON v.url_id = u.id
LEFT JOIN 
	browsers ON browsers.id = v.browser_id
CROSS JOIN (
  SELECT count(*) AS total
  FROM visits
  JOIN urls ON urls.id = visits.url_id
  WHERE urls.id = @url_id
) AS total_visits
WHERE
	u.author_id = @author_id
AND
    u.id = @url_id
GROUP BY
    browsers.name, 
    browsers.version, 
    browsers.platform, 
    browsers.mobile, 
    total_visits.total
ORDER BY
    visit_count DESC;

-- name: UniqueVisitCount :one
SELECT 
	count(*)
FROM
	visits
WHERE
	url_id = @url_id
GROUP BY 
	url_id, ip_address;


-- name: TotalVisit :one
SELECT count(*)
FROM
	visits
WHERE
	url_id = @url_id;


-- name: VisitOverTime :many
SELECT
    date_trunc(@time_trunc, visited_at)::timestamp AS visit_date,
    COUNT(*) AS visit_count
FROM
    visits
WHERE
    url_id = @url_id
    AND visited_at BETWEEN @start_date AND @end_date
GROUP BY
    visit_date
ORDER BY
    visit_date;

-- name: ReferrerDistribution :many
SELECT
    v.referrer AS source,
    count(v.id) AS click_count,
    (count(v.id) * 100.0 / total_visits.total)::float AS percentage
FROM
    visits v
JOIN
    urls u ON v.url_id = u.id
CROSS JOIN (
    SELECT count(*) AS total
    FROM visits
    WHERE visits.url_id = @url_id
) AS total_visits
WHERE
    u.author_id = @author_id
AND
    u.id = @url_id
AND 
    v.referrer IS NOT NULL AND v.referrer != '' -- Exclude empty or null referrers
GROUP BY
    v.referrer, total_visits.total
ORDER BY
    click_count DESC;