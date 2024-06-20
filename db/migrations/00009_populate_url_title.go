package migrations

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pressly/goose/v3"

	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/services"
)

func init() {
	goose.AddMigrationContext(
		upPopulateURLTitle,
		downPopulateURLTitle,
	)
}

type url struct {
	ID  int64
	URL string
}

func upPopulateURLTitle(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, "SELECT id, long_url from urls")
	if err != nil {
		return err
	}

	urls := []url{}
	for rows.Next() {
		var (
			ID      int64
			longURL string
		)
		if err := rows.Scan(&ID, &longURL); err != nil {
			return err
		}
		urls = append(urls, url{ID: ID, URL: longURL})
	}
	rows.Close()

	for _, u := range urls {
		log.Info("processing url", "url", u.URL, "id", u.ID)

		title := services.ExtractTitle(u.URL)
		title = strings.TrimSpace(title)
		title = title[:min(len(title), 500)]

		if title != "" {
			log.Info("title found", "title", title, "url", u.ID, "id", u.ID)

			_, err := tx.ExecContext(ctx, "UPDATE urls SET title = $1 WHERE id = $2", title, u.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func downPopulateURLTitle(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "UPDATE urls SET title = ''")
	return err
}
