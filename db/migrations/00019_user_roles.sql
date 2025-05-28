-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role    TEXT NOT NULL 
                 CHECK (role IN ('admin', 'user'))
                 DEFAULT 'user',
    PRIMARY KEY (user_id, role),
    FOREIGN KEY (user_id) REFERENCES users(guid) ON DELETE CASCADE
);

CREATE INDEX ON user_roles(role);
CREATE INDEX ON user_roles(user_id, role);


INSERT INTO user_roles(user_id, "role")
SELECT users.guid, 'user'
FROM users;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS user_roles_role_idx;
DROP INDEX IF EXISTS user_roles_user_id_role_idx;

DROP TABLE user_roles;
-- +goose StatementEnd