-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS session;

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;

CREATE TABLE IF NOT EXISTS session (
    key TEXT PRIMARY KEY,
    data BYTEA,
    expiry INTEGER
);
-- +goose StatementEnd
