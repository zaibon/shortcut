package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/urfave/cli/v2"

	_ "github.com/joho/godotenv/autoload"

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

	GeoIPBucket string
	GeoIPDBFile string

	ForceDev bool

	GoogleOauthClientID string
	GoogleOauthSecret   string
}

func (c config) DBString() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s timezone=UTC sslmode=disable", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

var c config

func main() {
	log.SetupLogger(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	app := &cli.App{
		Name:  "shortcut",
		Usage: "shortcut, your friendly shortening service",

		Flags: serverFlags,
		Action: func(ctx *cli.Context) error {
			return listenSignals(
				context.Background(), c,
				runServer,
				os.Interrupt, syscall.SIGTERM)
		},

		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "start the server",
				Args:  false,
				Action: func(ctx *cli.Context) error {
					return listenSignals(
						context.Background(), c,
						runServer,
						os.Interrupt, syscall.SIGTERM)
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
