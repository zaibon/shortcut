package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/templates"
	"github.com/zaibon/shortcut/templates/components"
)

type URLService interface {
	Shorten(ctx context.Context, url string, title string, userID domain.ID) (string, error)
	List(ctx context.Context, authorID domain.ID, search string) ([]domain.URLStat, error)
	Delete(ctx context.Context, urlID, authorID domain.ID) error

	Expand(ctx context.Context, short string) (domain.URL, error)
	ExtractTitle(url string) string

	Get(ctx context.Context, authorID domain.ID, slug string) (domain.URL, error)
	StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (domain.URLStat, error)

	ClickOverTime(ctx context.Context, urlID domain.ID, period domain.Period, timeRange string) ([]domain.TimeSeriesData, error)

	TrackRedirect(ctx context.Context, urlID domain.ID, r *http.Request) error

	CountMonthlyURL(ctx context.Context, authorID domain.ID) (int64, error)
	CountMonthlyVisit(ctx context.Context, authorID domain.ID) (int64, error)
}

type Handler struct {
	htmx *htmx.HTMX
	svc  URLService
}

func NewURLHandlers(shortURL URLService) *Handler {
	return &Handler{
		htmx: htmx.New(),
		svc:  shortURL,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.index)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)

		r.Post("/shorten", h.shorten)

		r.With(middleware.PaginateParams).Get("/urls", h.myLinks)
		r.With(middleware.PaginateParams).Get("/urls-sort", h.urlSort)
		r.With(middleware.PaginateParams).Get("/urls-search", h.urlSearch)

		r.Get("/urls/{slug}", h.linkDetail)
		r.Get("/urls/{id}/clicks", h.clickChart)
		r.Delete("/urls/{id}", h.deleteURL)
	})

	r.Get("/{shortID}", h.redirect)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	templates.IndexPage().Render(r.Context(), w)
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)
	ctx := r.Context()
	url := r.FormValue("url")
	title := r.FormValue("title")

	if errs := validateURL(url); len(errs) > 0 {
		addFlash(w, r, errs["long_url"].Error(), flashTypeError)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(ctx)
	if user == nil { // this should not happen
		htmx.Redirect("/auth")
		return
	}

	count, err := h.svc.CountMonthlyURL(ctx, user.ID)
	if err != nil {
		log.Error("failed to count monthly url", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if count >= domain.FreePlanLimit {
		addFlash(w, r, "You have reached your monthly limit of URLs", flashTypeError)
		// We return 200 OK so HTMX swaps the content (which will be empty/hidden or we could render a specific error state)
		// Or we can just return and let the flash message speak.
		// Since result-container has empty:hidden, sending empty string hides it.
		return
	}

	short, err := h.svc.Shorten(ctx, url, title, user.ID)
	if err != nil {
		log.Error("failed to shorten url", slog.Any("error", err))
		addFlash(w, r, "Failed to shorten URL\nSomething wrong happened, try again.", flashTypeError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addFlash(w, r, fmt.Sprintf("URL shortened to %s", short), flashTypeInfo)

	components.ShortenURL(short, url).Render(ctx, w)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "shortID")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	url, err := h.svc.Expand(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			templates.NotFoundPage().Render(r.Context(), w)
			return
		}

		log.Error("failed to expand url", slog.Any("error", err))
		http.Error(w, "failed to expand url", http.StatusInternalServerError)
		return
	}

	go func() {
		if err := h.svc.TrackRedirect(context.Background(), url.ID, r); err != nil {
			log.Error("failed to track redirect", slog.Any("error", err))
		}
	}()

	w.Header().Set("X-Robots-Tag", "noindex")
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
