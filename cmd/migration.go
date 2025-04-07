package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v2"

	"github.com/zaibon/shortcut/db"
)

var migrationScan = []cli.Flag{
	&cli.StringFlag{
		Name:        "db",
		Usage:       "database connection string",
		Value:       "postgres://localhost:4532",
		EnvVars:     []string{"SHORTCUT_DB"},
		Destination: &c.DBConnString,
	},
}

func runMigration(cCtx *cli.Context, c config) error {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(c.DBConnString)
	if err != nil {
		return fmt.Errorf("unable to parse database connection string: %v", err)
	}

	log.Printf("connecting to database: %s\n", c.SafeDBString())
	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	// defer dbPool.Close()

	dbConn := stdlib.OpenDBFromPool(dbPool)
	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	// defer dbConn.Close()

	return db.MigrateCmd(ctx, dbConn, cCtx.Args().First(), cCtx.Args().Tail()...)
}
