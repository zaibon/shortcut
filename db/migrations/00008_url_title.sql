-- +goose Up
-- +goose StatementBegin

ALTER TABLE urls ADD COLUMN title TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE urls DROP COLUMN title;
-- +goose StatementEnd
