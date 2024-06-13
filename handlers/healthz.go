package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type HealthzHandlers struct {
	db *sql.DB
}

func NewHealtzHandlers(db *sql.DB) *HealthzHandlers {
	return &HealthzHandlers{db: db}
}

func (h *HealthzHandlers) Routes(r chi.Router) {
	r.Get("/healthz", h.healthz)
}

func (h *HealthzHandlers) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	err := h.db.PingContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Write([]byte("ok"))
	w.WriteHeader(http.StatusOK)
}
