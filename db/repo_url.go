package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
)

type urlStore struct {
	db datastore.Querier
}

func NewURLStore(db *sql.DB) *urlStore {
	return &urlStore{
		db: datastore.New(db),
	}
}

func (s *urlStore) Add(ctx context.Context, shortURL, longURL string, authorID int64) (int64, error) {
	url, err := s.db.AddShortURL(ctx, datastore.AddShortURLParams{
		ShortUrl: shortURL,
		LongUrl:  longURL,
		AuthorID: sql.NullInt64{
			Int64: authorID,
			Valid: true,
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to add shorten url: %w", err)
	}
	return url.ID, nil
}

func (s *urlStore) List(ctx context.Context, authorID int64) ([]datastore.Url, error) {
	rows, err := s.db.ListShortURLs(ctx, sql.NullInt64{
		Int64: authorID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list shorten urls: %w", err)
	}
	return rows, nil
}

func (s *urlStore) Get(ctx context.Context, shortID string) (datastore.Url, error) {
	url, err := s.db.GetShortURL(ctx, shortID)
	if err != nil {
		return datastore.Url{}, fmt.Errorf("failed to get shorten url: %w", err)
	}

	return url, nil
}

func (s urlStore) TrackRedirect(ctx context.Context, urlID int64, ipAddress, userAgent string) error {
	if err := s.db.TrackRedirect(ctx, datastore.TrackRedirectParams{
		UrlID: sql.NullInt64{
			Int64: urlID,
			Valid: urlID != 0,
		},
		IpAddress: sql.NullString{
			String: ipAddress,
			Valid:  ipAddress != "",
		},
		UserAgent: sql.NullString{
			String: userAgent,
			Valid:  userAgent != "",
		},
	}); err != nil {
		return fmt.Errorf("failed to track redirect: %v", err)
	}

	return nil
}

func (s urlStore) Statistics(ctx context.Context, authorID int64) ([]datastore.ListStatisticsRow, error) {
	rows, err := s.db.ListStatistics(ctx, sql.NullInt64{
		Int64: authorID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list statistics: %w", err)
	}

	return rows, err
}
