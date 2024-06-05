package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zaibon/shortcut/views"
)

func (h *Handler) statistics(w http.ResponseWriter, r *http.Request) {
	urls, err := h.svc.Statistics(r.Context(), 1) //TODO
	if err != nil {
		h.log.Error("failed to get statistics", slog.Any("error", err))
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	data := views.StatisticsPageData{}
	for _, url := range urls {
		data.Stats = append(data.Stats, views.URLStat{
			Short:    url.Short,
			Long:     url.Long,
			NrVisite: strconv.Itoa(url.NrVisited),
		})
	}
	Render(r.Context(), w, views.StatisticsPage(data))
}
