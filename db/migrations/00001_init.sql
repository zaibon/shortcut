-- +goose Up
-- +goose StatementBegin
-- Table to store user information (author)
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table to store the URL mappings
CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_url TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    author_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id)
);

-- Table to store visit statistics
CREATE TABLE IF NOT EXISTS visits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_id INTEGER,
    visited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (url_id) REFERENCES urls(id)
);

-- Index to improve query performance for visits by url_id
CREATE INDEX IF NOT EXISTS idx_visits_url_id ON visits(url_id);

-- Index to improve query performance for urls by short_url
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS visits;
-- +goose StatementEnd
