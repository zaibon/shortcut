//go:build dev

package db

import (
	"os"
)

var migrationsFS = os.DirFS("./migrations")
