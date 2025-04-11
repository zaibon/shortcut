package static

import (
	"net/http"

	_ "embed"
)

//go:embed favicon/favicon.ico
var Favicon []byte

func FaviconHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.WriteHeader(http.StatusOK)
		w.Write(Favicon) //nolint:errcheck
	})
}
