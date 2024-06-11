//go:build dev

package static

import (
	"net/http"
)

var FileSystem http.FileSystem

func init() {
	FileSystem = http.Dir("./static")
}
