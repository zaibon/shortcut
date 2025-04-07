-- +goose Up
-- +goose StatementBegin
ALTER TABLE oauth2_state
ADD COLUMN provider text NOT NULL DEFAULT 'google'; 

CREATE INDEX idx_oauth2_state_state ON oauth2_state (state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE oauth2_state DROP COLUMN provider;

DROP INDEX idx_oauth2_state_state;
-- +goose StatementEnd
