-- +goose Up
-- +goose StatementBegin
CREATE TABLE oauth2_state (
    state VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expire_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '10 minutes'
);

ALTER TABLE users ADD COLUMN is_oauth BOOLEAN DEFAULT FALSE;
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_salt DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE oauth2_state;

ALTER TABLE users DROP COLUMN is_oauth;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
ALTER TABLE users ALTER COLUMN password_salt SET NOT NULL;
-- +goose StatementEnd
