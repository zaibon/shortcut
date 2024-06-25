package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/components"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/handlers/toast"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/views"
)

type ShortURLService interface {
	Shorten(ctx context.Context, url string, userID domain.ID) (string, error)
	List(ctx context.Context, authorID domain.ID) ([]domain.URL, error)
	Expand(ctx context.Context, short string) (domain.URL, error)

	StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (domain.URLStat, error)

	TrackRedirect(ctx context.Context, urlID domain.ID, r *http.Request) error
}

type Handler struct {
	svc ShortURLService
}

func NewURLHandlers(shortURL ShortURLService) *Handler {
	return &Handler{
		svc: shortURL,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.index)
	r.Get("/favicon.ico", h.favicon)
	r.Post("/shorten-url", h.shorten)
	r.Get("/{shortID}", h.redirect)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)
		r.Get("/links", h.myLinks)
		// r.Get("/links/{slug}", h.urlStatDetail)
		// r.Get("/stats-table", h.statsTable)

		r.Get("/statistics/clicks/{slug}", h.numberClicks)
	})
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	data := views.IndexPageData{}
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

	user := middleware.UserFromContext(ctx)
	if user == nil { // this should not happen
		HXRedirect(ctx, w, "/login")
		return
	}

	short, err := h.svc.Shorten(ctx, url, user.ID)
	if err != nil {
		log.Error("failed to shorten url", slog.Any("error", err))
		toast.Danger(w, "Failed to shorten URL", "Something wrong happened, try again.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	toast.Success(w, "Success", fmt.Sprintf("URL shortened to %s", short), `<a href="/links"></a>`)
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
		log.Error("failed to expand url", slog.Any("error", err))
		http.Error(w, "failed to expand url", http.StatusInternalServerError)
		return
	}

	go func() {
		if err := h.svc.TrackRedirect(context.Background(), url.ID, r); err != nil {
			log.Error("failed to track redirect", slog.Any("error", err))
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
