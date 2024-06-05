package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/components"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/views"
)

type ShortURLService interface {
	Shorten(ctx context.Context, url string) (string, error)
	List(ctx context.Context, authorID int64) ([]string, error)
	Expand(ctx context.Context, short string) (domain.URL, error)

	TrackRedirect(ctx context.Context, urlID int64, r *http.Request) error
}

type Handler struct {
	svc ShortURLService
	log *slog.Logger
}

func NewHandler(shortURL ShortURLService, log *slog.Logger) *Handler {
	return &Handler{
		svc: shortURL,
		log: log,
	}
}

func (h *Handler) Routes(r *chi.Mux) {
	r.Get("/", h.index)
	r.Get("/favicon.ico", h.favicon)
	r.Post("/shorten-url", h.shorten)
	r.Get("/{shortID}", h.redirect)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	data := views.IndexPageData{}

	urls, err := h.svc.List(r.Context(), 1)
	if err != nil {
		h.log.Error("failed to list shorten urls", slog.Any("error", err))
		Render(r.Context(), w, views.IndexPage(data))
		return
	}

	data.URLs = urls
	Render(r.Context(), w, views.IndexPage(data))
}

func (h *Handler) favicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	url := r.FormValue("long_url")

	if errs := validateURL(url); len(errs) > 0 {
		Render(r.Context(), w, components.IndexForm(components.FormData{
			URL:    url,
			Errors: errs,
		}))
		return
	}

	short, err := h.svc.Shorten(ctx, url)
	if err != nil {
		h.log.Error("failed to shorten url", slog.Any("error", err))
		//TODO: show toast
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Render(r.Context(), w, components.IndexForm(components.FormData{}))
	Render(r.Context(), w, components.AddedURL(short))
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "shortID")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	url, err := h.svc.Expand(r.Context(), id)
	if err != nil {
		h.log.Error("failed to expand url", slog.Any("error", err))
		http.Error(w, "failed to expand url", http.StatusInternalServerError)
		return
	}

	go func() {
		if err := h.svc.TrackRedirect(context.Background(), url.ID, r); err != nil {
			h.log.Error("failed to track redirect", slog.Any("error", err))
		}
	}()

	http.Redirect(w, r, url.Long, http.StatusMovedPermanently)
}

func validateURL(url string) map[string]error {
	errs := make(map[string]error)

	if url == "" {
		errs["long_url"] = fmt.Errorf("URL is required")
		return errs
	}

	if !IsValidURL(url) {
		errs["long_url"] = fmt.Errorf("invalid URL")
		return errs
	}

	return errs
}
