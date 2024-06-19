package main

import (
	"context"
	"fmt"
	"net/http"

	"gitea.com/go-chi/session"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v2"

	_ "gitea.com/go-chi/session/postgres"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/zaibon/shortcut/db"
	"github.com/zaibon/shortcut/handlers"
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
	// &cli.StringFlag{
	// 	Name:        "redis",
	// 	Usage:       "configuration string for redis server",
	// 	Value:       "network=tcp,addr=:6379,db=0,pool_size=100,idle_timeout=180,prefix=session;",
	// 	EnvVars:     []string{"SHORTCUT_REDIS"},
	// 	Destination: &c.Redis,
	// },
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

// runServer is the entry point of the application. It sets up the HTTP router, configures the database connection,
// applies any necessary database migrations, creates the URL shortening service, and registers the request handlers.
func runServer(c config) error {
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, c.DBString())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// databases
	urlStore := db.NewURLStore(dbPool)
	userStore := db.NewUserStore(dbPool)

	// services
	shortURL := services.NewShortURL(urlStore, c.redirectURL())
	userService := services.NewUser(userStore)

	// HTTP handlers
	urlHandlers := handlers.NewURLHandlers(shortURL)
	userHandlers := handlers.NewUsersHandler(userService)
	healthzHandlers := handlers.NewHealtzHandlers(stdlib.OpenDBFromPool(dbPool))

	fs := http.FileServer(static.FileSystem)
	server := chi.NewRouter()
	// no middlewares
	server.Group(func(r chi.Router) {
		r.Handle("/static/*", http.StripPrefix("/static/", fs))
	})

	// normal HTTP middlewares
	server.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Logger)
		r.Use(chiMiddleware.Recoverer)
		r.Use(chiMiddleware.RealIP)
		r.Use(session.Sessioner(
			session.Options{
				Provider:       "postgres",
				ProviderConfig: c.DBString(),
				CookieName:     "shortcut_session",
				Secure:         c.TLS,
				SameSite:       http.SameSiteLaxMode,
				IDLength:       32,
			},
		))
		r.Use(middleware.UserContext)

		// HTTP Routing
		urlHandlers.Routes(r)
		userHandlers.Routes(r)
		healthzHandlers.Routes(r)
	})

	listenAddr := fmt.Sprintf(":%d", c.Port)
	fmt.Printf("Server is running on %s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, server); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

func (c config) redirectURL() string {
	if c.TLS {
		return fmt.Sprintf("https://%s", c.Domain)
	} else {
		return fmt.Sprintf("http://%s", c.Domain)
	}
}
