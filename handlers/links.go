package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zaibon/shortcut/components"
	"github.com/zaibon/shortcut/handlers/toast"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/views"
)

const timeFormat = "Mon, 02 Jan 06 at 15:04:05"

func (h *Handler) myLinks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		HXRedirect(r.Context(), w, "/")
		return
	}

	archived := r.URL.Query().Get("archived")
	showArchived := archived == "true"

	urls, err := h.svc.List(r.Context(), user.ID, showArchived)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	data := views.MyLinkPageData{
		URLs:         []views.URLStat{},
		Autoreload:   r.URL.Query().Get("autoreload") == "true",
		ShowArchived: showArchived,
	}
	for _, url := range urls {
		data.URLs = append(data.URLs, views.URLStat{
			Title:      url.Title,
			Slug:       url.Slug,
			Short:      url.Short,
			Long:       url.Long,
			IsArchived: url.IsArchived,
			CreatedAt:  url.CreatedAt.Format(timeFormat),
		})
	}

	Render(r.Context(), w, views.MyLinksPage(data))
}

// titleEdit show the edit form for the title of the link on the detail page
func (h *Handler) titleEdit(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil { // this should not happen
		HXRedirect(r.Context(), w, "/login")
		return
	}

	slug := chi.URLParam(r, "slug")
	details, err := h.svc.Get(r.Context(), user.ID, slug)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	Render(r.Context(), w, components.EditTitleForm(details.Title, details.Slug))
}

// updateTitle update the title of the link
func (h *Handler) updateTitle(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil { // this should not happen
		HXRedirect(r.Context(), w, "/login")
		return
	}

	slug := chi.URLParam(r, "slug")
	if err := r.ParseForm(); err != nil {
		log.Error("failed to parse form", slog.Any("error", err))
		http.Error(w, "failed to update title", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	url, err := h.svc.UpdateTitle(r.Context(), user.ID, slug, title)
	if err != nil {
		log.Error("failed to update title", slog.Any("error", err))
		toast.Danger(w, "Failed to update title", "Something wrong happened, try again.")
		http.Error(w, "failed to update title", http.StatusInternalServerError)
		return
	}

	toast.Success(w, "Title updated", "")
	w.WriteHeader(http.StatusOK)
	Render(r.Context(), w, components.TitleDetail(title, slug, url.IsArchived))
}

func (h *Handler) archiveURL(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.ArchiveURL(r.Context(), user.ID, slug); err != nil {
		log.Error("failed to archive url", slog.Any("error", err))
		http.Error(w, "failed to archive url", http.StatusInternalServerError)
		return
	}
	toast.Success(w, "Success", "URL archived", "")
}

func (h *Handler) unarchiveURL(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.UnarchiveURL(r.Context(), user.ID, slug); err != nil {
		log.Error("failed to archive url", slog.Any("error", err))
		http.Error(w, "failed to archive url", http.StatusInternalServerError)
		return
	}

	toast.Success(w, "Success", "URL unarchived", "")
}
