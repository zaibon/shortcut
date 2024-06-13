-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS visit_locations (
    visit_id INTEGER NOT NULL,
    address TEXT,
    country_code CHAR(2),
    country_name TEXT,
    subdivision TEXT,
    continent TEXT,
    city_name TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    source TEXT,

    FOREIGN KEY (visit_id) REFERENCES visits (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS visit_locations;
-- +goose StatementEnd
