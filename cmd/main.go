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

	SentryDSN string

	SessionLifetime int

	GoogleWebRiskAPIKey string
}

func (c config) SafeDBString() string {
	u, _ := url.Parse(c.DBConnString)
	return u.Redacted()
}

// missingStripeConfig returns the names of any required Stripe settings that are
// empty. Billing (checkout, webhooks, customer portal) cannot work without them.
func (c config) missingStripeConfig() []string {
	var missing []string
	for name, v := range map[string]string{
		"stripe-key":             c.StripeKey,
		"stripe-pub-key":         c.StripePubKey,
		"stripe-endpoint-secret": c.StripeEndpointSecret,
	} {
		if v == "" {
			missing = append(missing, name)
		}
	}
	return missing
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
			{
				Name:  "scan-urls",
				Usage: "bulk scans existing shortened URLs against Google Web Risk API",
				Action: func(ctx *cli.Context) error {
					return runScanURLs(ctx, c)
				},
				Flags: scanURLFlags,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
