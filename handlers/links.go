package handlers

import (
	"log/slog"
	"net/http"

	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/views"
)

func (h *Handler) myLinks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		HXRedirect(r.Context(), w, "/")
		return
	}

	urls, err := h.svc.List(r.Context(), user.ID)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	data := views.MyLinkPageData{
		URLs:       []views.URLStat{},
		Autoreload: r.URL.Query().Get("autoreload") == "true",
	}
	for _, url := range urls {
		data.URLs = append(data.URLs, views.URLStat{
			Title:     url.Title,
			Slug:      url.Slug,
			Short:     url.Short,
			Long:      url.Long,
			CreatedAt: url.CreatedAt.Format("Mon, 02 Jan 06 at 15:04:05"),
		})
	}

	Render(r.Context(), w, views.MyLinksPage(data))
}
