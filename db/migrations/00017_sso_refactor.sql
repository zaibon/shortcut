-- +goose Up
-- +goose StatementBegin
-- Remove password-related columns from the users table
ALTER TABLE users
DROP COLUMN password,
DROP COLUMN password_salt;

-- Add updated_at column
ALTER TABLE users
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Create a trigger to update the updated_at column
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- Create the user_providers table
CREATE TABLE user_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL, -- Unique ID from the provider
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, provider_user_id),
    FOREIGN KEY (user_id) REFERENCES users(guid) ON DELETE CASCADE
);

-- Create a trigger to update the updated_at column
CREATE TRIGGER update_user_providers_updated_at
BEFORE UPDATE ON user_providers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop the user_providers table
DROP TABLE user_providers;

-- Remove updated_at column
ALTER TABLE users
DROP COLUMN updated_at;

-- Re-add password-related columns (if you need to revert)
ALTER TABLE users
ADD COLUMN password BYTEA,
ADD COLUMN password_salt BYTEA;

-- Drop the trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_user_providers_updated_at ON user_providers;
-- +goose StatementEnd
