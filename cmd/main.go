package main

import (
	"context"
	"log/slog"
	"net/url"
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

	DBConnString string

	GeoIPBucket string
	GeoIPDBFile string

	ForceDev bool

	GoogleOauthClientID string
	GoogleOauthSecret   string

	GithubOauthClientID string
	GithubOauthSecret   string

	StripeKey            string
	StripePubKey         string
	StripeEndpointSecret string
}

func (c config) SafeDBString() string {
	u, _ := url.Parse(c.DBConnString)
	return u.Redacted()
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
