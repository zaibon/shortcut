package handlers

import (
	"net/http"

	"github.com/zaibon/shortcut/templates"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	templates.NotFoundPage().Render(r.Context(), w)
}
