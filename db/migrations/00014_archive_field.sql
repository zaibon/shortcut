-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD is_archived BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN is_archived;
-- +goose StatementEnd
