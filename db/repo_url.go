package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/services/geoip"
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

func (s urlStore) TrackRedirect(ctx context.Context, urlID int64, ipAddress, userAgent string) (datastore.Visit, error) {
	visit, err := s.db.TrackRedirect(ctx, datastore.TrackRedirectParams{
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
	})
	if err != nil {
		return datastore.Visit{}, fmt.Errorf("failed to track redirect: %v", err)
	}
	return visit, nil
}

func (s urlStore) InsertVisitLocation(ctx context.Context, visitID int64, loc geoip.IPLocation) error {
	_, err := s.db.InsertVisitLocation(ctx, datastore.InsertVisitLocationParams{
		VisitID: visitID,
		Address: sql.NullString{
			String: loc.Address,
			Valid:  loc.Address != "",
		},
		CountryCode: nil,
		CountryName: sql.NullString{
			String: loc.CountryName,
			Valid:  loc.CountryName != "",
		},
		Subdivision: sql.NullString{
			String: loc.Subdivision,
			Valid:  loc.Subdivision != "",
		},
		Continent: sql.NullString{
			String: loc.Continent,
			Valid:  loc.Continent != "",
		},
		CityName: sql.NullString{
			String: loc.CityName,
			Valid:  loc.CityName != "",
		},
		Latitude: sql.NullFloat64{
			Float64: loc.Latitude,
			Valid:   loc.Latitude != 0,
		},
		Longitude: sql.NullFloat64{
			Float64: loc.Longitude,
			Valid:   loc.Longitude != 0,
		},
		Source: sql.NullString{
			String: loc.Source,
			Valid:  loc.Source != "",
		},
	})
	return err
}

func (s urlStore) Statistics(ctx context.Context, authorID int64) ([]datastore.ListStatisticsPerAuthorRow, error) {
	rows, err := s.db.ListStatisticsPerAuthor(ctx, sql.NullInt64{
		Int64: authorID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list statistics: %w", err)
	}

	return rows, err
}
