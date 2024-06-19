//go:build !dev

package static

import (
	"embed"
	"net/http"
)

//go:embed img js
var staticAssets embed.FS

var FileSystem http.FileSystem

func init() {
	FileSystem = http.FS(staticAssets)
}
