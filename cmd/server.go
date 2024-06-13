package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"gitea.com/go-chi/session"
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	_ "gitea.com/go-chi/session/redis"
	_ "github.com/mattn/go-sqlite3"

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
	&cli.StringFlag{
		Name:        "redis",
		Usage:       "configuration string for redis server",
		Value:       "network=tcp,addr=:6379,db=0,pool_size=100,idle_timeout=180,prefix=session;",
		EnvVars:     []string{"SHORTCUT_REDIS"},
		Destination: &c.Redis,
	},
	&cli.StringFlag{
		Name:        "db",
		Usage:       "path to the sqlite database file.",
		Value:       "shortcut.db",
		EnvVars:     []string{"SHORTCUT_DB"},
		Destination: &c.DBPath,
	},
}

// runServer is the entry point of the application. It sets up the HTTP router, configures the database connection,
// applies any necessary database migrations, creates the URL shortening service, and registers the request handlers.
func runServer(c config) error {
	dbConn, err := sql.Open("sqlite3", c.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer dbConn.Close()

	// databases
	urlStore := db.NewURLStore(dbConn)
	userStore := db.NewUserStore(dbConn)

	// services
	shortURL := services.NewShortURL(urlStore, c.redirectURL())
	userService := services.NewUser(userStore)

	// HTTP handlers
	urlHandlers := handlers.NewURLHandlers(shortURL)
	userHandlers := handlers.NewUsersHandler(userService)
	healthzHandlers := handlers.NewHealtzHandlers(dbConn)

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
				Provider:       "redis",
				ProviderConfig: c.Redis,
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
