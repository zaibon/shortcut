-- +goose Up
-- +goose StatementBegin
DELETE FROM urls WHERE created_at IS NULL;
ALTER TABLE urls ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE urls ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE urls ALTER COLUMN author_id SET NOT NULL;
ALTER TABLE urls ADD CONSTRAINT urls_unique_id UNIQUE (id);
SELECT setval('urls_id_seq', max(id), true) FROM urls;

UPDATE users SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD CONSTRAINT users_unique_id UNIQUE (id);
SELECT setval('users_id_seq', max(id), true) FROM users;

DELETE FROM visits WHERE visited_at IS NULL;
ALTER TABLE visits ALTER COLUMN visited_at SET NOT NULL;
ALTER TABLE visits ALTER COLUMN url_id SET NOT NULL;
ALTER TABLE visits ALTER COLUMN visited_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE visits ADD CONSTRAINT visits_unique_id UNIQUE (id);
SELECT setval('visits_id_seq', max(id), true) FROM visits;
-- -- +goose StatementEnd

-- -- +goose Down
-- -- +goose StatementBegin

ALTER TABLE urls ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE urls ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE urls ALTER COLUMN author_id DROP NOT NULL;
ALTER TABLE urls DROP CONSTRAINT IF EXISTS urls_unique_id;

ALTER TABLE users ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE users ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_unique_id;

ALTER TABLE visits ALTER COLUMN visited_at DROP NOT NULL;
ALTER TABLE visits ALTER COLUMN url_id DROP NOT NULL;
ALTER TABLE visits ALTER COLUMN visited_at DROP DEFAULT;
ALTER TABLE visits DROP CONSTRAINT IF EXISTS visits_unique_id;
-- +goose StatementEnd
