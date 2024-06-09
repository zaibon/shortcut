//go:build !dev

package db

import (
	"embed"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS
