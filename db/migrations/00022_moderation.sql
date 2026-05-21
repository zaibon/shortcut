-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS moderation_flags (
    id SERIAL PRIMARY KEY,
    url_id INTEGER NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    risk_score INTEGER NOT NULL DEFAULT 0,
    threat_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('flagged', 'approved', 'rejected')) DEFAULT 'flagged',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by UUID REFERENCES users(guid) ON DELETE SET NULL
);
CREATE INDEX ON moderation_flags(status);
CREATE INDEX ON moderation_flags(url_id);
CREATE INDEX ON moderation_flags(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS moderation_flags;
-- +goose StatementEnd
