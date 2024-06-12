package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gitea.com/go-chi/session"
	"github.com/go-chi/chi/v5"

	_ "gitea.com/go-chi/session/redis"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/zaibon/shortcut/db"
	"github.com/zaibon/shortcut/handlers"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/services"
	"github.com/zaibon/shortcut/static"
)

type config struct {
	Domain string
	Port   int

	Redis string
}

// main is the entry point of the application. It sets up the HTTP router, configures the database connection,
// applies any necessary database migrations, creates the URL shortening service, and registers the request handlers.
// The server is then started and listens on port 3333.
func main() {
	c := config{}
	flag.StringVar(&c.Domain, "domain", "localhost", "domain to use for shortened URLs")
	flag.IntVar(&c.Port, "port", 8080, "port to listen to")
	flag.StringVar(&c.Redis, "redis", "network=tcp,addr=:6379,db=0,pool_size=100,idle_timeout=180,prefix=session;", "configuration string for redis server")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	dbConn, err := sql.Open("sqlite3", "./shortcut.db")
	if err != nil {
		log.Error("failed to open database", slog.Any("error", err))
		return
	}
	defer dbConn.Close()

	if err := db.Migrate(dbConn); err != nil {
		log.Error("failed to apply migrations", slog.Any("error", err))
		return
	}

	// databases
	urlStore := db.NewURLStore(dbConn)
	userStore := db.NewUserStore(dbConn)

	// services
	shortURL := services.NewShortURL(urlStore, fmt.Sprintf("http://%s:%d", c.Domain, c.Port))
	userService := services.NewUser(userStore)

	// HTTP handlers
	urlHandlers := handlers.NewURLHandlers(shortURL, log)
	userHandlers := handlers.NewUsersHandler(userService)

	fs := http.FileServer(static.FileSystem)
	server := chi.NewRouter()
	// no middlewares
	server.Group(func(r chi.Router) {
		r.Handle("/static/*", http.StripPrefix("/static/", fs))
	})

	// normal HTTP middlewares
	r := server.Group(func(r chi.Router) {
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
	})

	// HTTP Routing
	urlHandlers.Routes(r)
	userHandlers.Routes(r)

	listenAddr := fmt.Sprintf(":%d", c.Port)
	fmt.Printf("Server is running on %s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, r); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "err", err)
	}
}
