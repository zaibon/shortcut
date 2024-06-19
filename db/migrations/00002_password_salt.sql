-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN password_salt BYTEA DEFAULT '' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN password_salt;
-- +goose StatementEnd
