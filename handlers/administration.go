package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/services"
	"github.com/zaibon/shortcut/templates/admin"
)

type AdministrationHandlers struct {
	service *services.Administration
}

func NewAdministrationHandlers(administrationService *services.Administration) *AdministrationHandlers {
	return &AdministrationHandlers{
		service: administrationService,
	}
}

func (h *AdministrationHandlers) Routes(r chi.Router, pool *pgxpool.Pool) {
	r.Group(func(r chi.Router) {

		db := datastore.New(pool)

		r.Use(middleware.Authenticated)
		r.Use(middleware.IsAdmin(db))
		r.Use(middleware.PaginateParams)

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/overview", http.StatusFound)
		})
		r.Get("/admin/overview", h.overview)
		r.Get("/admin/users", h.users)
		r.Get("/admin/users/{guid}", h.userDetail)
		r.Get("/admin/urls", h.urls)
		r.Get("/admin/urls/{slug}", h.urlDetail)
		r.Get("/admin/urls/{id}/edit", h.editURL)
		r.Post("/admin/urls/{id}", h.updateURL)
		r.Delete("/admin/urls/{id}", h.deleteURL)
		r.Patch("/admin/urls/{id}/status", h.toggleURLStatus)
		r.Patch("/admin/users/{guid}/status", h.toggleUserSuspension)
		r.Patch("/admin/users/{guid}/urls/status", h.toggleUserURLsStatus)
		r.Get("/admin/analytics", h.analytics)
		r.Get("/admin/settings", h.settings)
	})
}

func (h *AdministrationHandlers) userDetail(w http.ResponseWriter, r *http.Request) {
	guidStr := chi.URLParam(r, "guid")
	guid, err := domain.ParseGUID(guidStr)
	if err != nil {
		http.Error(w, "Invalid User GUID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(r.Context(), guid)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	urls, err := h.service.GetUserURLs(r.Context(), guid)
	if err != nil {
		http.Error(w, "Failed to retrieve user URLs", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab: "users",
	}

	admin.UserDetail(data, user, urls).Render(r.Context(), w)
}

func (h *AdministrationHandlers) editURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	url, err := h.service.GetURL(r.Context(), domain.ID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve URL", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab: "urls",
	}

	admin.URLEdit(data, url).Render(r.Context(), w)
}

func (h *AdministrationHandlers) updateURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	longURL := r.FormValue("long_url")

	if err := h.service.UpdateURL(r.Context(), domain.ID(id), title, longURL); err != nil {
		http.Error(w, "Failed to update URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/urls", http.StatusFound)
}

func (h *AdministrationHandlers) urlDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	stats, err := h.service.GetURLStats(r.Context(), slug)
	if err != nil {
		http.Error(w, "Failed to retrieve URL stats", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab: "urls",
	}

	admin.URLDetail(data, stats).Render(r.Context(), w)
}

func (h *AdministrationHandlers) deleteURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteURL(r.Context(), domain.ID(id)); err != nil {
		http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdministrationHandlers) toggleURLStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	isArchivedStr := r.FormValue("is_archived")
	isArchived := isArchivedStr == "true"
	isActiveStr := r.FormValue("is_active")
	isActive := isActiveStr == "true"

	if err := h.service.ToggleURLStatus(r.Context(), domain.ID(id), isArchived, isActive); err != nil {
		http.Error(w, "Failed to update URL status", http.StatusInternalServerError)
		return
	}

	target := r.Header.Get("HX-Target")

	// If we are on the detail page, update the header
	if target == "url-header" {
		url, err := h.service.GetURL(r.Context(), domain.ID(id))
		if err != nil {
			http.Error(w, "Failed to fetch updated URL", http.StatusInternalServerError)
			return
		}
		stats, err := h.service.GetURLStats(r.Context(), url.Slug)
		if err != nil {
			http.Error(w, "Failed to fetch updated URL stats", http.StatusInternalServerError)
			return
		}
		admin.URLDetailHeader(stats).Render(r.Context(), w)
		return
	}

	// Otherwise, we are likely on the list page, update the row
	urls, err := h.service.ListURLs(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch URLs", http.StatusInternalServerError)
		return
	}

	var updatedURL *domain.AdminURL
	for _, u := range urls {
		if u.ID == domain.ID(id) {
			updatedURL = &u
			break
		}
	}

	if updatedURL == nil {
		http.Error(w, "Updated URL not found", http.StatusInternalServerError)
		return
	}

	admin.URLRow(*updatedURL).Render(r.Context(), w)
}

func (h *AdministrationHandlers) toggleUserSuspension(w http.ResponseWriter, r *http.Request) {
	guidStr := chi.URLParam(r, "guid")
	guid, err := domain.ParseGUID(guidStr)
	if err != nil {
		http.Error(w, "Invalid User GUID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	isSuspendedStr := r.FormValue("is_suspended")
	isSuspended := isSuspendedStr == "true"

	if err := h.service.ToggleUserSuspension(r.Context(), guid, isSuspended); err != nil {
		http.Error(w, "Failed to update user suspension status", http.StatusInternalServerError)
		return
	}

	updatedUser, err := h.service.GetUser(r.Context(), guid)
	if err != nil {
		http.Error(w, "Failed to fetch updated user", http.StatusInternalServerError)
		return
	}

	target := r.Header.Get("HX-Target")
	if target == "user-header" {
		admin.UserDetailHeader(updatedUser).Render(r.Context(), w)
		return
	}

	admin.UserRow(updatedUser).Render(r.Context(), w)
}

func (h *AdministrationHandlers) toggleUserURLsStatus(w http.ResponseWriter, r *http.Request) {
	guidStr := chi.URLParam(r, "guid")
	guid, err := domain.ParseGUID(guidStr)
	if err != nil {
		http.Error(w, "Invalid User GUID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	isActiveStr := r.FormValue("is_active")
	isActive := isActiveStr == "true"

	if err := h.service.ToggleUserURLsStatus(r.Context(), guid, isActive); err != nil {
		http.Error(w, "Failed to update user URLs status", http.StatusInternalServerError)
		return
	}

	// We redirect to the user detail page to refresh everything
	// Alternatively, we could re-render the whole UserDetail template and let HTMX swap the body
	// or specific parts. Since we are changing ALL URLs, the list needs full refresh.
	// Simple approach: reload the page or return the full detail view if called via HTMX with body swap.
	// Better yet, just reuse userDetail logic but without the wrapper if needed, or just redirect.
	// If using hx-boost or similar, a redirect works. If using hx-patch, we can return the new content.
	// Let's reuse userDetail logic to return the full page content for replacement.

	user, err := h.service.GetUser(r.Context(), guid)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	urls, err := h.service.GetUserURLs(r.Context(), guid)
	if err != nil {
		http.Error(w, "Failed to retrieve user URLs", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab: "users",
	}

	admin.UserDetail(data, user, urls).Render(r.Context(), w)
}

func (h *AdministrationHandlers) overview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.GetOverviewStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve overview statistics", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab:      "overview",
		Overview: *overview,
	}

	admin.OverviewTab(data).Render(r.Context(), w)
}

func (h *AdministrationHandlers) users(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	paginatePrams := middleware.GetPaginationParams(r.Context())
	paginationLinks := middleware.GeneratePaginationLinks(paginatePrams, len(users))
	users = middleware.Paginate(users, paginatePrams)

	data := admin.AdminDashboardData{
		Tab:        "users",
		Users:      users,
		Pagination: paginationLinks,
	}

	admin.UsersTab(data).Render(r.Context(), w)
}

func (h *AdministrationHandlers) urls(w http.ResponseWriter, r *http.Request) {
	urls, err := h.service.ListURLs(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve URLs", http.StatusInternalServerError)
		return
	}

	paginatePrams := middleware.GetPaginationParams(r.Context())
	paginationLinks := middleware.GeneratePaginationLinks(paginatePrams, len(urls))
	urls = middleware.Paginate(urls, paginatePrams)

	data := admin.AdminDashboardData{
		Tab:        "urls",
		URLs:       urls,
		Pagination: paginationLinks,
	}

	admin.URLsTab(data).Render(r.Context(), w)
}
func (h *AdministrationHandlers) analytics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAnalyticsStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve analytics", http.StatusInternalServerError)
		return
	}

	data := admin.AdminDashboardData{
		Tab:       "analytics",
		Analytics: *stats,
	}

	admin.AnalyticsTab(data).Render(r.Context(), w)
}
func (h *AdministrationHandlers) settings(w http.ResponseWriter, r *http.Request) {
	data := admin.AdminDashboardData{
		Tab: "settings",
	}

	admin.SettingsTab(data).Render(r.Context(), w)
}
