package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitea.com/go-chi/session"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v2"

	_ "gitea.com/go-chi/session/postgres"

	sentryhttp "github.com/getsentry/sentry-go/http"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/zaibon/shortcut/db"
	"github.com/zaibon/shortcut/env"
	"github.com/zaibon/shortcut/handlers"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/services"
	"github.com/zaibon/shortcut/static"
)

var serverFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:        "tls",
		Usage:       "generate redirect URL using HTTPS",
		Value:       true,
		EnvVars:     []string{"SHORTCUT_TLS"},
		Destination: &c.TLS,
	},
	&cli.StringFlag{
		Name:        "domain",
		Usage:       "domain to use for shortened URLs",
		Value:       "localhost:8080",
		EnvVars:     []string{"SHORTCUT_DOMAIN"},
		Destination: &c.Domain,
	},
	&cli.IntFlag{
		Name:        "port",
		Usage:       "port to listen to",
		Value:       8080,
		EnvVars:     []string{"SHORTCUT_PORT"},
		Destination: &c.Port,
	},
	&cli.StringFlag{
		Name:        "db",
		Usage:       "database connection string",
		Value:       "postgres://localhost:4532",
		EnvVars:     []string{"SHORTCUT_DB"},
		Destination: &c.DBConnString,
	},
	&cli.BoolFlag{
		Name:        "dev",
		Usage:       "Force dev mode",
		Value:       false,
		EnvVars:     []string{"SHORTCUT_DEV_MODE"},
		Destination: &c.ForceDev,
	},
	&cli.StringFlag{
		Name:        "google-oauth-client-id",
		Usage:       "Google OAuth client ID",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_GOOGLE_OAUTH_CLIENT_ID"},
		Destination: &c.GoogleOauthClientID,
	},
	&cli.StringFlag{
		Name:        "google-oauth-secret",
		Usage:       "Google OAuth secret",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_GOOGLE_OAUTH_SECRET"},
		Destination: &c.GoogleOauthSecret,
	},
	&cli.StringFlag{
		Name:        "github-oauth-client-id",
		Usage:       "Github OAuth client ID",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_GITHUB_OAUTH_CLIENT_ID"},
		Destination: &c.GithubOauthClientID,
	},
	&cli.StringFlag{
		Name:        "github-oauth-secret",
		Usage:       "Github OAuth secret",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_GITHUB_OAUTH_SECRET"},
		Destination: &c.GithubOauthSecret,
	},
	&cli.StringFlag{
		Name:        "stripe-key",
		Usage:       "Stripe secret key",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_STRIPE_KEY"},
		Destination: &c.StripeKey,
	},
	&cli.StringFlag{
		Name:        "stripe-pub-key",
		Usage:       "Stripe public key",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_STRIPE_PUB_KEY"},
		Destination: &c.StripePubKey,
	},
	&cli.StringFlag{
		Name:        "stripe-endpoint-secret",
		Usage:       "Stripe endpoint secret",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_STRIPE_ENDPOINT_SECRET"},
		Destination: &c.StripeEndpointSecret,
	},
	&cli.StringFlag{
		Name:        "sentry-dsn",
		Usage:       "Sentry DSN for error tracking",
		Value:       "",
		EnvVars:     []string{"SHORTCUT_SENTRY_DSN"},
		Destination: &c.SentryDSN,
	},
}

func listenSignals(ctx context.Context, c config, f func(context.Context, config) error, sig ...os.Signal) error {
	ctx, cancel := context.WithCancel(ctx)

	cSig := make(chan os.Signal, 1)
	signal.Notify(cSig, os.Interrupt, syscall.SIGTERM)

	cErr := make(chan error, 1)
	go func() {
		cErr <- f(ctx, c)
	}()

	<-cSig
	cancel()
	log.Info("shutting down server")

	select {
	case <-time.After(time.Second * 5):
		log.Error("server shutdown timeout")
		return nil

	case err := <-cErr:
		log.Info("server shutdown complete")
		return err
	}
}

// runServer is the entry point of the application. It sets up the HTTP router, configures the database connection,
// applies any necessary database migrations, creates the URL shortening service, and registers the request handlers.
func runServer(ctx context.Context, c config) error {
	dbPool, err := pgxpool.New(ctx, c.DBConnString)
	if err != nil {
		return fmt.Errorf("unable to connect to database %s: %v", c.SafeDBString(), err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping the database %s: %v", c.SafeDBString(), err)
	}

	// databases
	urlStore := db.NewURLStore(dbPool)
	userStore := db.NewUserStore(dbPool)
	subscriptionStore := db.NewRepoSubscription(dbPool)

	// services
	urlService := services.NewURL(urlStore, c.redirectURL())
	userService := services.NewUser(userStore, c.Domain, c.TLS,
		c.GoogleOauthClientID, c.GoogleOauthSecret,
		c.GithubOauthClientID, c.GithubOauthSecret,
	)
	stripeService := services.NewStripe(c.StripeKey, subscriptionStore, c.Domain, c.TLS)
	adminService := services.NewAdministrationService(dbPool, c.redirectURL())

	// setup Sentry for error tracking
	setupSentry(c)

	// HTTP handlers
	urlHandlers := handlers.NewURLHandlers(urlService, stripeService)
	userHandlers := handlers.NewUsersHandler(userService, stripeService, urlService, c.StripePubKey)
	healthzHandlers := handlers.NewHealtzHandlers(stdlib.OpenDBFromPool(dbPool))
	subscriptionHandlers := handlers.NewSubscriptionHandlers(c.StripeKey, c.StripeEndpointSecret, stripeService, urlService)
	adminHandlers := handlers.NewAdministrationHandlers(adminService)

	fs := http.FileServer(static.FileSystem)
	server := chi.NewRouter()

	// Create an instance of sentryhttp
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         time.Second * 2,
	})
	server.Use(sentryHandler.Handle)
	server.NotFound(handlers.NotFound)

	// no middlewares
	server.Group(func(r chi.Router) {
		r.Handle("/static/*", http.StripPrefix("/static/", fs))
		r.Handle("/robot.txt", static.RobotsHandler())
		r.Handle("/favicon.ico", static.FaviconHandler())
		r.Handle("/sitemap.xml", static.SitemapHandler())
	})

	// normal HTTP middlewares
	server.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Logger)
		r.Use(chiMiddleware.Recoverer)
		r.Use(chiMiddleware.RealIP)
		r.Use(session.Sessioner(
			session.Options{
				Provider:       "postgres",
				ProviderConfig: c.DBConnString,
				CookieName:     "shortcut_session",
				Secure:         false,
				SameSite:       http.SameSiteLaxMode,
				IDLength:       32,
			},
		))
		r.Use(middleware.UserContext)
		r.Use(middleware.SentryMiddleware)

		// HTTP Routing
		urlHandlers.Routes(r)
		userHandlers.Routes(r)
		healthzHandlers.Routes(r)
		subscriptionHandlers.Routes(r)
		adminHandlers.Routes(r, dbPool)
	})

	listenAddr := fmt.Sprintf(":%d", c.Port)
	log.Info("Server is running", "addr", listenAddr, "env", env.Name())
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: server,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	log.Info("shutting down server gracefully...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "err", err)
		return err
	}

	return nil
}

func setupSentry(c config) {
	if c.SentryDSN == "" {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: c.SentryDSN,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for tracing.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
		SampleRate:       1.0,
		EnableTracing:    true,
		SendDefaultPII:   false,
		Environment:      env.Name(),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	log.Info("Sentry initialized", "dsn", c.SentryDSN, "env", env.Name())
}

func (c config) redirectURL() string {
	if c.TLS {
		return fmt.Sprintf("https://%s", c.Domain)
	} else {
		return fmt.Sprintf("http://%s", c.Domain)
	}
}
