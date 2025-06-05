package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"

	"github.com/zaibon/shortcut/db/datastore"
)

//TODO zaibon: This is noe needed anymore, might need to clean it up

func init() {
	goose.AddMigrationNoTxContext(
		pgxMigrateFunc(upPopulateGeoip),
		downPopulateGeoip,
	)
}

func upPopulateGeoip(ctx context.Context, store datastore.Querier) error {
	return nil
}

func downPopulateGeoip(ctx context.Context, db *sql.DB) error {
	return nil
}
