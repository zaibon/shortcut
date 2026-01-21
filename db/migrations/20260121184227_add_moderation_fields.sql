-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN is_suspended BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN is_active;
ALTER TABLE users DROP COLUMN is_suspended;
-- +goose StatementEnd