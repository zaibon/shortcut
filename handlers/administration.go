package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zaibon/shortcut/db/datastore"
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
		r.Get("/admin/urls", h.urls)
		r.Get("/admin/analytics", h.analytics)
		r.Get("/admin/settings", h.settings)
	})
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
