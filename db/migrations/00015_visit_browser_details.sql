-- +goose Up
-- +goose StatementBegin
CREATE TABLE browsers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    platform TEXT NOT NULL,    
    mobile BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_browser UNIQUE (name, version, platform, mobile)
);

ALTER TABLE visits
ADD COLUMN browser_id UUID REFERENCES browsers(id) ON DELETE SET NULL;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE visits
DROP COLUMN browser_id;

DROP TABLE browsers;
-- +goose StatementEnd
