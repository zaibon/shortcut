package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/zaibon/shortcut/db/datastore"
)

func pgxMigrateFunc(migration func(ctx context.Context, store datastore.Querier) error) goose.GoMigrationNoTxContext {
	return func(ctx context.Context, db *sql.DB) error {
		conn, err := db.Conn(ctx)
		if err != nil {
			return err
		}

		err = conn.Raw(func(driverConn any) error {
			conn := driverConn.(*stdlib.Conn).Conn() // conn is a *pgx.Conn

			store := datastore.New(conn)
			tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return err
			}
			store = store.WithTx(tx)

			if err := migration(ctx, store); err != nil {
				if err := tx.Rollback(ctx); err != nil {
					slog.Error("failed to rollback transaction", "err", err)
				}
				return err
			}

			if err := tx.Commit(ctx); err != nil {
				slog.Error("failed to commit transaction", "err", err)
				return err
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to run raw query: %w", err)
		}

		return nil
	}
}
