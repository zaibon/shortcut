package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

func Render(ctx context.Context, w http.ResponseWriter, component templ.Component) {
	if err := component.Render(ctx, w); err != nil {
		slog.Error("failed to render short url", slog.Any("error", err))
		http.Error(w, "failed to render short url", http.StatusInternalServerError)
	}
}
