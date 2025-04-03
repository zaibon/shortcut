-- +goose Up
-- +goose StatementBegin
-- Table to store user information (author)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table to store the URL mappings
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    author_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id) 
                            ON DELETE CASCADE ON UPDATE CASCADE
                            DEFERRABLE INITIALLY DEFERRED
);

-- Table to store visit statistics
CREATE TABLE IF NOT EXISTS visits (
    id SERIAL PRIMARY KEY,
    url_id INTEGER,
    visited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (url_id) REFERENCES urls(id)
                         ON DELETE CASCADE ON UPDATE CASCADE
                         DEFERRABLE INITIALLY DEFERRED
);

-- Index to improve query performance for visits by url_id
CREATE INDEX IF NOT EXISTS idx_visits_url_id ON visits(url_id);

-- Index to improve query performance for urls by short_url
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS visits;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
