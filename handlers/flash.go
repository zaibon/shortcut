package handlers

import (
	"net/http"

	"github.com/donseba/go-htmx"
)

type flashType string

const (
	flashTypeInfo  flashType = "info"
	flashTypeError flashType = "error"
)

func addFlash(w http.ResponseWriter, r *http.Request, message string, flashType flashType) {
	htmx.New().NewHandler(w, r).
		TriggerCustom("flash", message, map[string]any{
			"type": string(flashType),
		})
}
