-- name: IsAdmin :one
SELECT
    EXISTS (
        SELECT 1
        FROM users
        JOIN user_roles
          ON users.guid = user_roles.user_id
        WHERE guid = @guid AND role = 'admin'
    ) AS is_admin;

-- name: AdminListUsers :many
SELECT
    u.*,
    COALESCE(s.stripe_product_name, 'Free') AS plan_name,
    COUNT(DISTINCT urls.id) AS url_count,
    COUNT(DISTINCT v.id) AS click_count,
    CASE 
        WHEN u.is_suspended THEN 'suspended'
        ELSE 'active'
    END AS user_status
FROM
    users u
LEFT JOIN
    customer c ON u.guid = c.user_id
LEFT JOIN
    subscription s ON c.user_id = s.customer_id
LEFT JOIN
    urls ON u.id = urls.author_id
LEFT JOIN
    visits v ON urls.id = v.url_id
WHERE
    (sqlc.narg('search')::text IS NULL OR u.username ILIKE '%' || sqlc.narg('search') || '%' OR u.email ILIKE '%' || sqlc.narg('search') || '%')
    AND (sqlc.narg('is_suspended')::boolean IS NULL OR u.is_suspended = sqlc.narg('is_suspended'))
    AND (sqlc.narg('plan')::text IS NULL OR COALESCE(s.stripe_product_name, 'Free') = sqlc.narg('plan'))
    AND (sqlc.narg('created_after')::timestamp IS NULL OR u.created_at >= sqlc.narg('created_after'))
GROUP BY
    u.id, u.username, u.email, u.created_at, s.stripe_product_name, s.status
ORDER BY
    u.created_at DESC;

-- name: AdminGetUser :one
SELECT
    u.*,
    COALESCE(s.stripe_product_name, 'Free') AS plan_name,
    COUNT(DISTINCT urls.id) AS url_count,
    COUNT(DISTINCT v.id) AS click_count,
    CASE 
        WHEN u.is_suspended THEN 'suspended'
        ELSE 'active'
    END AS user_status
FROM
    users u
LEFT JOIN
    customer c ON u.guid = c.user_id
LEFT JOIN
    subscription s ON c.user_id = s.customer_id
LEFT JOIN
    urls ON u.id = urls.author_id
LEFT JOIN
    visits v ON urls.id = v.url_id
WHERE u.guid = @guid
GROUP BY
    u.id, u.username, u.email, u.created_at, s.stripe_product_name, s.status;


-- name: AdminGetOverviewStatistics :one
WITH monthly_user_activity AS (
    SELECT
        SUM(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) AND created_at < (date_trunc('month', CURRENT_DATE) + interval '1 month') THEN 1 ELSE 0 END) AS new_users_current_month,
        SUM(CASE WHEN created_at >= (date_trunc('month', CURRENT_DATE) - interval '1 month') AND created_at < date_trunc('month', CURRENT_DATE) THEN 1 ELSE 0 END) AS new_users_previous_month
    FROM users
), monthly_url_activity AS (
    SELECT
        SUM(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) AND created_at < (date_trunc('month', CURRENT_DATE) + interval '1 month') THEN 1 ELSE 0 END) AS new_urls_current_month,
        SUM(CASE WHEN created_at >= (date_trunc('month', CURRENT_DATE) - interval '1 month') AND created_at < date_trunc('month', CURRENT_DATE) THEN 1 ELSE 0 END) AS new_urls_previous_month
    FROM urls
), monthly_visit_activity AS (
    SELECT
        SUM(CASE WHEN visited_at >= date_trunc('month', CURRENT_DATE) AND visited_at < (date_trunc('month', CURRENT_DATE) + interval '1 month') THEN 1 ELSE 0 END) AS new_visits_current_month,
        SUM(CASE WHEN visited_at >= (date_trunc('month', CURRENT_DATE) - interval '1 month') AND visited_at < date_trunc('month', CURRENT_DATE) THEN 1 ELSE 0 END) AS new_visits_previous_month
    FROM visits
)
SELECT
    (SELECT COUNT(*) FROM users) AS total_users,
    (SELECT COUNT(*) FROM urls) AS total_urls,
    (SELECT COUNT(*) FROM visits) AS total_clicks,
    CASE
        WHEN mua.new_users_previous_month = 0 THEN
            CASE
                WHEN mua.new_users_current_month > 0 THEN 100.0 -- Indicate growth from zero
                ELSE 0.0 -- No change from zero
            END
        ELSE (CAST(mua.new_users_current_month AS REAL) - mua.new_users_previous_month) * 100.0 / mua.new_users_previous_month
    END AS total_users_variation,
    CASE
        WHEN mula.new_urls_previous_month = 0 THEN
            CASE
                WHEN mula.new_urls_current_month > 0 THEN 100.0
                ELSE 0.0
            END
        ELSE (CAST(mula.new_urls_current_month AS REAL) - mula.new_urls_previous_month) * 100.0 / mula.new_urls_previous_month
    END AS total_urls_variation,
    CASE
        WHEN mva.new_visits_previous_month = 0 THEN
            CASE
                WHEN mva.new_visits_current_month > 0 THEN 100.0
                ELSE 0.0
            END
        ELSE (CAST(mva.new_visits_current_month AS REAL) - mva.new_visits_previous_month) * 100.0 / mva.new_visits_previous_month
    END AS total_clicks_variation
FROM
    monthly_user_activity mua,
    monthly_url_activity mula,
    monthly_visit_activity mva;


-- name: AdminGetUserGrowth :many
WITH date_series AS (
    SELECT generate_series(
        (CURRENT_DATE - interval '29 days')::date, -- Start 30 days ago (inclusive of today)
        CURRENT_DATE::date,                        -- End today
        '1 day'::interval
    )::date AS day
)
SELECT
    ds.day ::timestamp AS "day",
    COALESCE(COUNT(u.created_at), 0)::bigint AS count
FROM
    date_series ds
LEFT JOIN
    users u ON date_trunc('day', u.created_at) = ds.day
GROUP BY
    ds.day
ORDER BY
    ds.day;




-- name: AdminGetURLCreationTrends :many
WITH date_series AS (
    SELECT generate_series(
        (CURRENT_DATE - interval '29 days')::date, -- Start 30 days ago (inclusive of today)
        CURRENT_DATE::date,                        -- End today
        '1 day'::interval
    )::date AS day
)
SELECT
    ds.day ::timestamp AS "day",
    COALESCE(COUNT(u.created_at), 0)::bigint  AS count
FROM
    date_series ds
LEFT JOIN
    urls u ON date_trunc('day', u.created_at) = ds.day
GROUP BY
    ds.day
ORDER BY
    ds.day;


-- name: AdminGetTotalUsersTrend :many
WITH date_series AS (
    SELECT generate_series(
        (CURRENT_DATE - interval '29 days')::date, -- Start 30 days ago (inclusive of today)
        CURRENT_DATE::date,                        -- End today
        '1 day'::interval
    )::date AS day
)
SELECT
    ds.day::timestamp AS "day",
    (
        SELECT COUNT(*)
        FROM users u
        WHERE date_trunc('day', u.created_at) <= ds.day
    )::bigint AS count
FROM
    date_series ds
ORDER BY
    ds.day;


-- name: AdminListURLSDetails :many
SELECT
    sqlc.embed(urls),
    sqlc.embed(users),
    users.username AS author_name,
    COUNT(visits.id) AS click_count
FROM
    urls
JOIN
    users ON urls.author_id = users.id
LEFT JOIN
    visits ON urls.id = visits.url_id
LEFT JOIN
    customer c ON users.guid = c.user_id
LEFT JOIN
    subscription s ON c.user_id = s.customer_id
WHERE
    (sqlc.narg('search')::text IS NULL OR urls.title ILIKE '%' || sqlc.narg('search') || '%' OR urls.short_url ILIKE '%' || sqlc.narg('search') || '%' OR urls.long_url ILIKE '%' || sqlc.narg('search') || '%')
    AND (sqlc.narg('is_active')::boolean IS NULL OR urls.is_active = sqlc.narg('is_active'))
    AND (sqlc.narg('plan')::text IS NULL OR COALESCE(s.stripe_product_name, 'Free') = sqlc.narg('plan'))
    AND (sqlc.narg('created_after')::timestamp IS NULL OR urls.created_at >= sqlc.narg('created_after'))
GROUP BY
    urls.id, users.username, users.id, s.stripe_product_name
HAVING
    (sqlc.narg('min_clicks')::int IS NULL OR COUNT(visits.id) >= sqlc.narg('min_clicks'))
    AND (sqlc.narg('max_clicks')::int IS NULL OR COUNT(visits.id) <= sqlc.narg('max_clicks'))
ORDER BY
    urls.created_at DESC;

-- name: AdminListUserURLs :many
SELECT
    sqlc.embed(urls),
    sqlc.embed(users),
    users.username AS author_name,
    COUNT(visits.id) AS click_count
FROM
    urls
JOIN
    users ON urls.author_id = users.id
LEFT JOIN
    visits ON urls.id = visits.url_id
WHERE
    users.guid = @guid
GROUP BY
    urls.id, users.username, users.id
ORDER BY
    urls.created_at DESC;

-- name: AdminGetDailyActiveVisitors :many
WITH date_series AS (
    SELECT generate_series(
        (CURRENT_DATE - interval '6 days')::date,
        CURRENT_DATE::date,
        '1 day'::interval
    )::date AS day
)
SELECT
    ds.day::timestamp AS "day",
    COUNT(DISTINCT v.ip_address)::bigint AS count
FROM
    date_series ds
LEFT JOIN
    visits v ON date_trunc('day', v.visited_at) = ds.day
GROUP BY
    ds.day
ORDER BY
    ds.day;

-- name: AdminGetTopReferrers :many
SELECT
    COALESCE(NULLIF(referrer, ''), 'Direct') AS source,
    COUNT(*)::bigint AS count
FROM
    visits
GROUP BY
    source
ORDER BY
    count DESC
LIMIT 5;

-- name: AdminGetTopURLs :many
SELECT
    u.short_url,
    u.long_url,
    COUNT(v.id)::bigint AS clicks
FROM
    urls u
JOIN
    visits v ON u.id = v.url_id
GROUP BY
    u.id
ORDER BY
    clicks DESC
LIMIT 5;

-- name: AdminGetGeoDistribution :many
SELECT
    COALESCE(vl.country_name, 'Unknown') AS country,
    COUNT(*)::bigint AS count
FROM
    visits v
JOIN
    visit_locations vl ON v.id = vl.visit_id
GROUP BY
    country
ORDER BY
    count DESC
LIMIT 5;

-- name: AdminDeleteURL :exec
DELETE FROM urls
WHERE id = @id;

-- name: AdminUpdateURLStatus :exec
UPDATE urls
SET is_archived = @is_archived,
    is_active = @is_active
WHERE id = @id;

-- name: AdminToggleUserURLs :exec
UPDATE urls
SET is_active = @is_active
FROM users
WHERE urls.author_id = users.id AND users.guid = @guid;

-- name: AdminUpdateURL :one
UPDATE urls
SET title = @title,
    long_url = @long_url
WHERE id = @id
RETURNING *;

-- name: AdminGetRecentActivity :many
SELECT
    'user_registered' AS type,
    u.username AS actor,
    u.email AS details,
    u.created_at AS occurred_at
FROM
    users u
UNION ALL
SELECT
    'url_created' AS type,
    u.username AS actor,
    url.short_url AS details,
    url.created_at AS occurred_at
FROM
    urls url
JOIN
    users u ON url.author_id = u.id
ORDER BY
    occurred_at DESC
LIMIT 10;
