package static

import (
	"embed"
	"net/http"

	"github.com/zaibon/shortcut/env"
)

//go:embed img js css favicon sitemap
var staticAssets embed.FS

var FileSystem http.FileSystem

func init() {
	if env.IsDev() {
		FileSystem = http.Dir("./static")
	} else {
		FileSystem = http.FS(staticAssets)
	}
}
