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
    'active' AS user_status
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
GROUP BY
    u.id, u.username, u.email, u.created_at, s.stripe_product_name, s.status
ORDER BY
    u.created_at DESC;


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
