package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zaibon/shortcut/db"
	"github.com/zaibon/shortcut/handlers"
	"github.com/zaibon/shortcut/services"
)

// main is the entry point of the application. It sets up the HTTP router, configures the database connection,
// applies any necessary database migrations, creates the URL shortening service, and registers the request handlers.
// The server is then started and listens on port 3333.
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

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

	store := db.NewURLStore(dbConn)
	shortURL := services.NewShortURL(store, "http://localhost:3333")
	handlers := handlers.NewHandler(shortURL, log)

	handlers.Routes(r)

	fmt.Println("Server is running on port 3333")
	if err := http.ListenAndServe(":3333", r); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "err", err)
	}
}
