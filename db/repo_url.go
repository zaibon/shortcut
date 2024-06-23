package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/services/geoip"
)

type urlStore struct {
	db datastore.Querier
}

func NewURLStore(db datastore.DBTX) *urlStore {
	return &urlStore{
		db: datastore.New(db),
	}
}

func (s *urlStore) Add(ctx context.Context, title, shortURL, longURL string, authorID domain.ID) (domain.ID, error) {
	url, err := s.db.AddShortURL(ctx, datastore.AddShortURLParams{
		Title:    title,
		ShortUrl: shortURL,
		LongUrl:  longURL,
		AuthorID: int32(authorID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to add shorten url: %w", err)
	}
	return domain.ID(url.ID), nil
}

func (s *urlStore) List(ctx context.Context, authorID domain.ID, sortBy domain.URLSortRequest) ([]datastore.Url, error) {
	rows, err := s.db.ListShortURLs(ctx, datastore.ListShortURLsParams{
		AuthorID: int32(authorID),
		SortBy:   sortBy.String(),
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

func (s urlStore) TrackRedirect(ctx context.Context, urlID domain.ID, ipAddress, userAgent string) (datastore.Visit, error) {
	visit, err := s.db.TrackRedirect(ctx, datastore.TrackRedirectParams{
		UrlID: int32(urlID),
		IpAddress: pgtype.Text{
			String: ipAddress,
			Valid:  ipAddress != "",
		},
		UserAgent: pgtype.Text{
			String: userAgent,
			Valid:  userAgent != "",
		},
	})
	if err != nil {
		return datastore.Visit{}, fmt.Errorf("failed to track redirect: %v", err)
	}
	return visit, nil
}

func (s urlStore) InsertVisitLocation(ctx context.Context, visitID domain.ID, loc geoip.IPLocation) error {
	_, err := s.db.InsertVisitLocation(ctx, datastore.InsertVisitLocationParams{
		VisitID: int32(visitID),
		Address: pgtype.Text{
			String: loc.Address,
			Valid:  loc.Address != "",
		},
		CountryCode: pgtype.Text{
			String: loc.CountryCode,
			Valid:  false,
		},
		CountryName: pgtype.Text{
			String: loc.CountryName,
			Valid:  loc.CountryName != "",
		},
		Subdivision: pgtype.Text{
			String: loc.Subdivision,
			Valid:  loc.Subdivision != "",
		},
		Continent: pgtype.Text{
			String: loc.Continent,
			Valid:  loc.Continent != "",
		},
		CityName: pgtype.Text{
			String: loc.CityName,
			Valid:  loc.CityName != "",
		},
		Latitude: pgtype.Float8{
			Float64: loc.Latitude,
			Valid:   loc.Latitude != 0,
		},
		Longitude: pgtype.Float8{
			Float64: loc.Longitude,
			Valid:   loc.Longitude != 0,
		},
		Source: pgtype.Text{
			String: loc.Source,
			Valid:  loc.Source != "",
		},
	})
	return err
}

func (s urlStore) Statistics(ctx context.Context, authorID domain.ID) ([]datastore.ListStatisticsPerAuthorRow, error) {
	rows, err := s.db.ListStatisticsPerAuthor(ctx, int32(authorID))
	if err != nil {
		return nil, fmt.Errorf("failed to list statistics: %w", err)
	}

	return rows, err
}

func (s urlStore) StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (domain.URLStat, error) {
	row, err := s.db.StatisticPerURL(ctx, datastore.StatisticPerURLParams{
		ShortUrl: slug,
		AuthorID: int32(authorID),
	})
	if err != nil {
		return domain.URLStat{}, fmt.Errorf("failed to get statistics: %w", err)
	}

	return domain.URLStat{
		URL: domain.URL{
			ID:        domain.ID(row.ID),
			Long:      row.LongUrl,
			Short:     row.ShortUrl,
			Slug:      slug,
			CreatedAt: row.CreatedAt.Time,
		},
		NrVisited: int(row.NrVisits),
	}, nil
}
