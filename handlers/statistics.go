package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/views"
)

func (h *Handler) statistics(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		HXRedirect(r.Context(), w, "/")
		return
	}

	urls, err := h.svc.Statistics(r.Context(), user.ID)
	if err != nil {
		h.log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	data := views.StatisticsPageData{
		Stats:      []views.URLStat{},
		Autoreload: r.URL.Query().Get("autoreload") == "true",
	}
	for _, url := range urls {
		data.Stats = append(data.Stats, views.URLStat{
			Short:    url.Short,
			Long:     url.Long,
			NrVisite: strconv.Itoa(url.NrVisited),
		})
	}

	Render(r.Context(), w, views.StatisticsPage(data))
}

func (h *Handler) statsTable(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		HXRedirect(r.Context(), w, "/")
		return
	}

	urls, err := h.svc.Statistics(r.Context(), user.ID)
	if err != nil {
		h.log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	data := views.StatisticsPageData{
		Autoreload: r.URL.Query().Get("autoreload") == "true",
	}
	for _, url := range urls {
		data.Stats = append(data.Stats, views.URLStat{
			Short:    url.Short,
			Long:     url.Long,
			NrVisite: strconv.Itoa(url.NrVisited),
		})
	}
	Render(r.Context(), w, views.StatTable(data))
}
