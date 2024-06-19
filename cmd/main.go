package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	_ "gitea.com/go-chi/session/redis"
	_ "github.com/mattn/go-sqlite3"

	"github.com/zaibon/shortcut/log"
)

type config struct {
	TLS    bool
	Domain string
	Port   int

	// Redis string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

func (c config) DBString() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

var c config

func main() {
	log.SetupLogger(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	if err := godotenv.Load(); err != nil {
		log.Error("unable to load .env file", "error", err)
		os.Exit(1)
	}

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
