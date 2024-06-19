-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS session (
    key TEXT PRIMARY KEY,
    data BYTEA,
    expiry INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session;
-- +goose StatementEnd
