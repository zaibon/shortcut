-- name: GetCustomerByStripeId :one
-- Description: Get customer by stripe id
SELECT * 
FROM customer
WHERE stripe_id = $1
LIMIT 1;

-- name: InsertCustomer :one
INSERT INTO customer (user_id, stripe_id)
VALUES ($1, $2)
RETURNING *;

-- name: InsertSubscription :one
INSERT INTO subscription (stripe_id, customer_id, stripe_price_id, stripe_product_name, status, quantity)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListCustomerSubscription :many
SELECT *
FROM subscription
WHERE customer_id = @customer_id
AND (sqlc.narg(status)::text IS NULL OR sqlc.narg(status) = status)
ORDER BY updated_at DESC;