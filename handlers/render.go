package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/zaibon/shortcut/services"
)

func Render(ctx context.Context, w http.ResponseWriter, component templ.Component) {
	if err := component.Render(ctx, w); err != nil {
		slog.Error("failed to render short url", slog.Any("error", err))
		http.Error(w, "failed to render short url", http.StatusInternalServerError)
	}
}

func HXRedirect(ctx context.Context, w http.ResponseWriter, href string) {
	w.Header().Set("HX-Redirect", href)
}

func ErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	errMap := map[int][]error{
		http.StatusUnprocessableEntity: {
			services.ErrInvalidCredentials,
			services.ErrUserNotFound,
		},
	}

	for status, errs := range errMap {
		for _, target := range errs {
			if errors.Is(err, target) {
				return status
			}
		}
	}

	return http.StatusInternalServerError
}
