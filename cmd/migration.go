package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v2"

	"github.com/zaibon/shortcut/db"
)

var migrationScan = []cli.Flag{
	&cli.StringFlag{
		Name:        "db-host",
		Usage:       "database host",
		Value:       "localhost",
		EnvVars:     []string{"SHORTCUT_DB_HOST"},
		Destination: &c.DBHost,
	},
	&cli.IntFlag{
		Name:        "db-port",
		Usage:       "database port",
		Value:       5432,
		EnvVars:     []string{"SHORTCUT_DB_PORT"},
		Destination: &c.DBPort,
	},
	&cli.StringFlag{
		Name:        "db-user",
		Usage:       "database user",
		Value:       "shortcut",
		EnvVars:     []string{"SHORTCUT_DB_USER"},
		Destination: &c.DBUser,
	},
	&cli.StringFlag{
		Name:        "db-password",
		Usage:       "database password",
		Value:       "shortcut",
		EnvVars:     []string{"SHORTCUT_DB_PASSWORD"},
		Destination: &c.DBPassword,
	},
	&cli.StringFlag{
		Name:        "db-name",
		Usage:       "database name",
		Value:       "shortcut",
		EnvVars:     []string{"SHORTCUT_DB_NAME"},
		Destination: &c.DBName,
	},
}

func runMigration(cCtx *cli.Context, c config) error {
	ctx := context.Background()

	dbPool, err := pgxpool.New(ctx, c.DBString())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	dbConn := stdlib.OpenDBFromPool(dbPool)
	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	return db.MigrateCmd(ctx, dbConn, cCtx.Args().First(), cCtx.Args().Tail()...)
}
