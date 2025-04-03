package handlers

// func (h *Handler) numberClicks(w http.ResponseWriter, r *http.Request) {
// 	user := middleware.UserFromContext(r.Context())
// 	if user == nil { // this should not happen
// 		HXRedirect(r.Context(), w, "/login")
// 		return
// 	}

// 	slug := chi.URLParam(r, "slug")

// 	url, err := h.svc.StatisticsDetail(r.Context(), user.ID, slug)
// 	if err != nil {
// 		log.Error("failed to get statistics", slog.Any("error", err))
// 		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	if url.NrVisited <= 0 {
// 		fmt.Fprintf(w, "no click")
// 		return
// 	} else if url.NrVisited == 1 {
// 		fmt.Fprintf(w, "1 click")
// 		return
// 	} else {
// 		fmt.Fprintf(w, "%d clicks", url.NrVisited)
// 	}
// }
