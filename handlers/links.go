package handlers

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/templates"
	"github.com/zaibon/shortcut/templates/components"
)

func Paginate[T any](s []T, q *PaginationQuery) []T {
	start := q.Offset()
	end := q.Offset() + q.Limit()
	switch {
	case start >= len(s):
		return s[:0]
	case end > len(s):
		return s[start:]
	default:
		return s[start:end]
	}
}

func (h *Handler) myLinks(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/auth")
		return
	}

	urls, err := h.svc.List(r.Context(), user.ID, "")
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	paginationLinks := middleware.GeneratePaginationLinks(pagination, len(urls))

	urls = middleware.Paginate(urls, pagination)
	urls = sortUrls(urls, "")

	if err := templates.URLSPage(urls, paginationLinks).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) urlSort(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/auth")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error("failed to parse form", slog.Any("error", err))
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	sortBy := r.FormValue("sort")

	urls, err := h.svc.List(r.Context(), user.ID, "")
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	paginationLinks := middleware.GeneratePaginationLinks(pagination, len(urls))

	urls = middleware.Paginate(urls, pagination)
	urls = sortUrls(urls, sortBy)

	if err := templates.URLList(urls, paginationLinks).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) urlSearch(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/auth")
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error("failed to parse form", slog.Any("error", err))
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	search := r.FormValue("search")

	urls, err := h.svc.List(r.Context(), user.ID, search)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	paginationLinks := middleware.GeneratePaginationLinks(pagination, len(urls))

	urls = middleware.Paginate(urls, pagination)

	if err := templates.URLList(urls, paginationLinks).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) deleteURL(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/auth")
		return
	}

	if user.IsSuspended {
		http.Error(w, "Account is suspended", http.StatusForbidden)
		return
	}

	sid := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(sid, 10, 32)
	if err != nil {
		log.Error("failed to parse url id", slog.Any("error", err))
		http.Error(w, "failed to parse url id", http.StatusBadRequest)
		return
	}

	urlID := domain.ID(id)
	if err := h.svc.Delete(r.Context(), urlID, user.ID); err != nil {
		log.Error("failed to delete url", slog.Any("error", err))
		http.Error(w, "failed to delete url", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) linkDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	user := middleware.UserFromContext(r.Context())

	url, err := h.svc.StatisticsDetail(r.Context(), user.ID, slug)
	if err != nil {
		log.Error("failed to get url", slog.Any("error", err))
		http.Error(w, "failed to get url", http.StatusInternalServerError)
		return
	}

	if err := templates.URLDetail(url).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) clickChart(w http.ResponseWriter, r *http.Request) {
	sID := chi.URLParam(r, "id")
	urlID, err := strconv.ParseInt(sID, 10, 32)
	if err != nil {
		log.Error("failed to parse url id", slog.Any("error", err))
		http.Error(w, "failed to parse url id", http.StatusBadRequest)
		return
	}

	now := time.Now()
	period := domain.Period{
		Until: time.Now(),
	}

	timeRange := r.URL.Query().Get("range")

	switch timeRange {
	case "day":
		period.Since = now.AddDate(0, 0, -1)
	case "week":
		period.Since = now.AddDate(0, 0, -7)
	case "month":
		period.Since = now.AddDate(0, -1, 0)
	default:
		period.Since = now.AddDate(0, 0, -7)
	}

	data, err := h.svc.ClickOverTime(r.Context(), domain.ID(urlID), period, timeRange)
	if err != nil {
		log.Error("failed to get click over time", slog.Any("error", err))
		http.Error(w, "failed to get click over time", http.StatusInternalServerError)
		return
	}

	if err := components.ChartData("visitOverTime", data).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))

	}
}

var (
	sortByNewest = func(i, j domain.URLStat) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	}
	sortByOldest = func(i, j domain.URLStat) int {
		return i.CreatedAt.Compare(j.CreatedAt)
	}
	compareInt = func(i, j int) int {
		return j - i
	}
	sortByMostClicks = func(i, j domain.URLStat) int {
		return compareInt(i.NrVisited, j.NrVisited)

	}
	sortByLeastClicks = func(i, j domain.URLStat) int {
		return compareInt(j.NrVisited, i.NrVisited)
	}
)

func sortUrls(urls []domain.URLStat, filter string) []domain.URLStat {
	switch filter {
	case "newest":
		slices.SortFunc(urls, sortByNewest)
	case "oldest":
		slices.SortFunc(urls, sortByOldest)
	case "most-clicked":
		slices.SortFunc(urls, sortByMostClicks)
	case "least-clicked":
		slices.SortFunc(urls, sortByLeastClicks)
	default:
		slices.SortFunc(urls, sortByNewest)

	}
	return urls
}
