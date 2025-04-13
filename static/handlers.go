package static

import (
	"net/http"

	_ "embed"
)

//go:embed favicon/favicon.ico
var favicon []byte

func FaviconHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.WriteHeader(http.StatusOK)
		w.Write(favicon) //nolint:errcheck
	})
}

//go:embed sitemap/sitemap.xml
var sitemap []byte

func SitemapHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write(sitemap) //nolint:errcheck
	})
}
