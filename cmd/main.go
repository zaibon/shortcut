package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	_ "gitea.com/go-chi/session/redis"
	_ "github.com/mattn/go-sqlite3"

	"github.com/zaibon/shortcut/log"
)

type config struct {
	TLS    bool
	Domain string
	Port   int

	Redis string

	DBPath       string
	MigrationDir string
}

var c config

func main() {
	log.SetupLogger(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	app := &cli.App{
		Name:  "shortcut",
		Usage: "shortcut, your friendly shortening service",

		Flags: serverFlags,
		Action: func(ctx *cli.Context) error {
			return runServer(c)
		},

		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "start the server",
				Args:  false,
				Action: func(ctx *cli.Context) error {
					return runServer(c)
				},
				Flags: serverFlags,
			},
			{
				Name:      "migrate",
				Usage:     "run DB migrations",
				Args:      true,
				ArgsUsage: "",
				Action: func(ctx *cli.Context) error {
					return runMigration(ctx, c)
				},
				Flags: migrationScan,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
