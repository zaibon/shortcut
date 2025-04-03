package handlers

import (
	"log/slog"
	"math/rand"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/templates"
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
		htmx.Redirect("/login")
		return
	}

	urls, err := h.svc.List(r.Context(), user.ID, "")
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// 	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	pagination.TotalRecords = len(urls)

	urls = middleware.Paginate(urls, &pagination)
	urls = sortUrls(urls, "")

	if err := templates.URLSPage(urls, pagination).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) urlSort(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/login")
		return
	}

	r.ParseForm()
	sortBy := r.FormValue("sort")

	urls, err := h.svc.List(r.Context(), user.ID, "")
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// 	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	pagination.TotalRecords = len(urls)

	urls = middleware.Paginate(urls, &pagination)
	urls = sortUrls(urls, sortBy)

	if err := templates.URLTable(urls, pagination).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) urlSearch(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/login")
		return
	}

	r.ParseForm()
	search := r.FormValue("search")

	urls, err := h.svc.List(r.Context(), user.ID, search)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	// Get pagination parameters from the context
	pagination := middleware.GetPaginationParams(r.Context())
	pagination.TotalRecords = len(urls)

	urls = middleware.Paginate(urls, &pagination)

	if err := templates.URLTable(urls, pagination).
		Render(r.Context(), w); err != nil {
		log.Error("failed to render page", slog.Any("error", err))
	}
}

func (h *Handler) deleteURL(w http.ResponseWriter, r *http.Request) {
	htmx := h.htmx.NewHandler(w, r)
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		htmx.Redirect("/login")
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

	// data := domain.URLStat{
	// 	URL:                  url,
	// 	UniqueVisitors:       url.NrVisited,
	// 	LocationDistribution: generateLocationDistribution(),
	// 	Referrers:            generateReferrers(),
	// 	Devices:              generateDevices(),
	// 	Browsers:             generateBrowsers(),
	// }

	if err := templates.URLDetail(url).
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

// titleEdit show the edit form for the title of the link on the detail page
// func (h *Handler) titleEdit(w http.ResponseWriter, r *http.Request) {
// 	user := middleware.UserFromContext(r.Context())
// 	if user == nil { // this should not happen
// 		HXRedirect(r.Context(), w, "/login")
// 		return
// 	}

// 	slug := chi.URLParam(r, "slug")
// 	details, err := h.svc.Get(r.Context(), user.ID, slug)
// 	if err != nil {
// 		log.Error("failed to get statistics", slog.Any("error", err))
// 		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
// 		return
// 	}

// 	Render(r.Context(), w, components.EditTitleForm(details.Title, details.Slug))
// }

// // updateTitle update the title of the link
// func (h *Handler) updateTitle(w http.ResponseWriter, r *http.Request) {
// 	user := middleware.UserFromContext(r.Context())
// 	if user == nil { // this should not happen
// 		HXRedirect(r.Context(), w, "/login")
// 		return
// 	}

// 	slug := chi.URLParam(r, "slug")
// 	if err := r.ParseForm(); err != nil {
// 		log.Error("failed to parse form", slog.Any("error", err))
// 		http.Error(w, "failed to update title", http.StatusInternalServerError)
// 		return
// 	}

// 	title := r.FormValue("title")
// 	if title == "" {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	url, err := h.svc.UpdateTitle(r.Context(), user.ID, slug, title)
// 	if err != nil {
// 		log.Error("failed to update title", slog.Any("error", err))
// 		toast.Danger(w, "Failed to update title", "Something wrong happened, try again.")
// 		http.Error(w, "failed to update title", http.StatusInternalServerError)
// 		return
// 	}

// 	toast.Success(w, "Title updated", "")
// 	w.WriteHeader(http.StatusOK)
// 	Render(r.Context(), w, components.TitleDetail(title, slug, url.IsArchived))
// }

// func (h *Handler) archiveURL(w http.ResponseWriter, r *http.Request) {
// 	slug := chi.URLParam(r, "slug")
// 	if slug == "" {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	user := middleware.UserFromContext(r.Context())
// 	if user == nil {
// 		http.Error(w, "unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	if err := h.svc.ArchiveURL(r.Context(), user.ID, slug); err != nil {
// 		log.Error("failed to archive url", slog.Any("error", err))
// 		http.Error(w, "failed to archive url", http.StatusInternalServerError)
// 		return
// 	}
// 	toast.Success(w, "Success", "URL archived", "")
// }

// func (h *Handler) unarchiveURL(w http.ResponseWriter, r *http.Request) {
// 	slug := chi.URLParam(r, "slug")
// 	if slug == "" {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	user := middleware.UserFromContext(r.Context())
// 	if user == nil {
// 		http.Error(w, "unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	if err := h.svc.UnarchiveURL(r.Context(), user.ID, slug); err != nil {
// 		log.Error("failed to archive url", slog.Any("error", err))
// 		http.Error(w, "failed to archive url", http.StatusInternalServerError)
// 		return
// 	}

// 	toast.Success(w, "Success", "URL unarchived", "")
// }

// generateReferrers creates sample referrer data.
func generateReferrers() []domain.Referrer {
	referrers := []string{"Direct / None", "Google", "Twitter", "Facebook", "LinkedIn", "Reddit", "Other"}
	referrerData := make([]domain.Referrer, len(referrers))
	totalClicks := 0

	for i, ref := range referrers {
		clicks := rand.Intn(500) + 50 // Generate clicks between 50 and 550
		referrerData[i] = domain.Referrer{
			Source:     ref,
			ClickCount: clicks,
		}
		totalClicks += clicks
	}

	for i := range referrerData {
		referrerData[i].Percentage = (float32(referrerData[i].ClickCount) / float32(totalClicks)) * 100
	}

	return referrerData
}
