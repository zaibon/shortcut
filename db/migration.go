package db

import (
	"context"
	"log"

	"github.com/pressly/goose/v3"

	_ "github.com/zaibon/shortcut/db/migrations"
	_ "modernc.org/sqlite" //TODO: replace with github.com/mattn/go-sqlite3 ?
)

func MigrateCmd(ctx context.Context, migrationDir, dbstring, command string, args ...string) {
	db, err := goose.OpenDBWithDriver("sqlite", dbstring)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	if err := goose.RunContext(ctx, command, db, migrationDir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
