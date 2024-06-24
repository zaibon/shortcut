-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN guid UUID NOT NULL DEFAULT gen_random_uuid();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN guid;
-- +goose StatementEnd
