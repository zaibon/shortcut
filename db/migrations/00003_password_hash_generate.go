package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/zaibon/shortcut/services/password"
)

func init() {
	goose.AddMigrationContext(upPasswordHashGenerate, downPasswordHashGenerate)
}

func upPasswordHashGenerate(ctx context.Context, tx *sql.Tx) error {
	hasher := password.DefaultArgon2iHasher()

	rows, err := tx.QueryContext(ctx, "SELECT id,password FROM users;")
	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			id     int64
			passwd string
		)

		if err := rows.Scan(&id, &passwd); err != nil {
			return err
		}

		saltedHash, err := hasher.Hash([]byte(passwd), nil)
		if err != nil {
			return err
		}

		result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET password = ?, password_salt = ?
		WHERE id = ?;`,
			saltedHash.Hash, saltedHash.Salt, id)
		if err != nil {
			return err
		}

		n, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if n != 1 {
			return fmt.Errorf("should have 1 row modified only, have %d", n)
		}
	}

	// This code is executed when the migration is applied.
	return nil
}

func downPasswordHashGenerate(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
