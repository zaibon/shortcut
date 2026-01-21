-- name: InsertUserOauth :one
INSERT INTO users (guid, username, email, is_oauth)
VALUES (@guid, @username, @email, true)
ON CONFLICT (email) DO UPDATE SET username = @username
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE users.email = @email
LIMIT 1;

-- name: GetUserByGUID :one
SELECT *
FROM users
WHERE users.guid = @guid
LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE guid = @guid;

-- name: InsertOauth2State :exec
INSERT INTO oauth2_state (state, provider)
VALUES (@state, @provider);

-- name: GetOauth2State :one
SELECT *
FROM oauth2_state
WHERE state = @state
AND expire_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: InsertUserProvider :one
INSERT INTO user_providers (user_id, provider, provider_user_id)
VALUES (@user_id, @provider, @provider_user_id)
RETURNING *;

-- name: GetUserProvider :one
SELECT *
FROM user_providers
WHERE user_id = @user_id
AND provider = @provider
LIMIT 1;

-- name: GetUserProviderByProviderUserId :one
SELECT *
FROM user_providers
WHERE provider = @provider
AND provider_user_id = @provider_user_id
LIMIT 1;

-- name: ListUserProviders :many
SELECT *
FROM user_providers
WHERE user_id = @user_id
ORDER BY created_at DESC;

-- name: UpdateUserSuspension :exec
UPDATE users
SET is_suspended = @is_suspended
WHERE guid = @guid;