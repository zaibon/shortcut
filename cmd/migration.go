package main

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/zaibon/shortcut/db"
)

var migrationScan = []cli.Flag{
	&cli.StringFlag{
		Name:        "db",
		Usage:       "path to the sqlite database file.",
		Value:       "shortcut.db",
		EnvVars:     []string{"SHORTCUT_DB"},
		Destination: &c.DBPath,
	},
	&cli.StringFlag{
		Name:        "migration-dir",
		Usage:       "directory containing the database migrations. If Specified the binary just apply migrations to the database and exit.",
		EnvVars:     []string{"SHORTCUT_MIGRATION_DIR"},
		Destination: &c.MigrationDir,
	},
}

func runMigration(cCtx *cli.Context, c config) error {
	ctx := context.Background()

	db.MigrateCmd(ctx, c.MigrationDir, c.DBPath, cCtx.Args().First(), cCtx.Args().Tail()...)
	return nil
}
