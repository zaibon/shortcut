package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (s *urlStore) List(ctx context.Context, authorID domain.ID, search string) ([]datastore.ListStatisticsPerAuthorRow, error) {
	rows, err := s.db.ListStatisticsPerAuthor(ctx, datastore.ListStatisticsPerAuthorParams{
		AuthorID: int32(authorID),
		Search:   search,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to list shorten urls: %w", err)
	}
	return rows, nil
}

func (s *urlStore) Delete(ctx context.Context, urlID, authorID domain.ID) error {
	return s.db.DeleteURL(ctx, datastore.DeleteURLParams{
		ID:       int32(urlID),
		AuthorID: int32(authorID),
	})
}

func (s *urlStore) Get(ctx context.Context, slug string) (datastore.Url, error) {
	url, err := s.db.GetShortURL(ctx, slug)
	if err != nil {
		return datastore.Url{}, fmt.Errorf("failed to get shorten url: %w", err)
	}

	return url, nil
}

func (s *urlStore) GetByID(ctx context.Context, urlID domain.ID) (datastore.Url, error) {
	url, err := s.db.GetByID(ctx, int32(urlID))
	if err != nil {
		return datastore.Url{}, fmt.Errorf("failed to get shorten url: %w", err)
	}

	return url, nil
}

func (s urlStore) TrackRedirect(ctx context.Context, urlID domain.ID, request domain.RequestInfo) (datastore.Visit, error) {

	browser, err := s.UpsertBrowser(ctx, request.Browser())
	if err != nil {
		return datastore.Visit{}, fmt.Errorf("upsert browser: %w", err)
	}

	visit, err := s.db.TrackRedirect(ctx, datastore.TrackRedirectParams{
		UrlID: int32(urlID),
		IpAddress: pgtype.Text{
			String: request.IpAddress(),
			Valid:  request.IpAddress() != "",
		},
		UserAgent: pgtype.Text{
			String: request.UserAgent(),
			Valid:  request.UserAgent() != "",
		},
		BrowserID: pgtype.UUID{
			Bytes: browser.ID,
			Valid: !browser.ID.IsNil(),
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
	rows, err := s.db.ListStatisticsPerAuthor(ctx, datastore.ListStatisticsPerAuthorParams{
		AuthorID: int32(authorID),
		Search:   "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list statistics: %w", err)
	}

	return rows, err
}

func (s urlStore) StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (datastore.StatisticPerURLRow, error) {
	row, err := s.db.StatisticPerURL(ctx, datastore.StatisticPerURLParams{
		ShortUrl: slug,
		AuthorID: int32(authorID),
	})
	if err != nil {
		return datastore.StatisticPerURLRow{}, fmt.Errorf("failed to get statistics: %w", err)
	}

	return row, nil
}

func (s urlStore) UpdateTitle(ctx context.Context, authorID domain.ID, slug, title string) (datastore.Url, error) {
	return s.db.UpdateTitle(ctx, datastore.UpdateTitleParams{
		Title:    title,
		ShortUrl: slug,
		AuthorID: int32(authorID),
	})
}

func (s urlStore) ArchiveURL(ctx context.Context, authorID domain.ID, slug string) error {
	return s.db.ArchiveURL(ctx, datastore.ArchiveURLParams{
		ShortUrl: slug,
		AuthorID: int32(authorID),
	})
}

func (s urlStore) UnarchiveURL(ctx context.Context, authorID domain.ID, slug string) error {
	return s.db.UnarchiveURL(ctx, datastore.UnarchiveURLParams{
		ShortUrl: slug,
		AuthorID: int32(authorID),
	})
}

func (a urlStore) CountMonthlyURL(ctx context.Context, authorID domain.ID) (int64, error) {
	return a.db.CountURLThisMonth(ctx, int32(authorID))
}

func (a urlStore) CountMonthlyVisit(ctx context.Context, authorID domain.ID) (int64, error) {
	return a.db.CountTotalVisitThisMonth(ctx, int32(authorID))
}

func (a urlStore) LocationDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.LocationDistributionRow, error) {
	return a.db.LocationDistribution(ctx, datastore.LocationDistributionParams{
		UrlID:    int32(urlID),
		AuthorID: int32(authorID),
	})
}

func (a urlStore) BrowserDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.BrowserDistributionRow, error) {
	return a.db.BrowserDistribution(ctx, datastore.BrowserDistributionParams{
		UrlID:    int32(urlID),
		AuthorID: int32(authorID),
	})
}

func (a urlStore) RefererDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.ReferrerDistributionRow, error) {
	return a.db.ReferrerDistribution(ctx, datastore.ReferrerDistributionParams{
		UrlID:    int32(urlID),
		AuthorID: int32(authorID),
	})
}

func (a urlStore) UniqueVisitCount(ctx context.Context, urlID domain.ID) (int64, error) {
	return a.db.UniqueVisitCount(ctx, int32(urlID))
}

func (a urlStore) TotalVisit(ctx context.Context, urlID domain.ID) (int64, error) {
	return a.db.TotalVisit(ctx, int32(urlID))
}

func (a urlStore) UpsertBrowser(ctx context.Context, browser domain.Browser) (domain.Browser, error) {
	row, err := a.db.UpsertBrowser(ctx, datastore.UpsertBrowserParams{
		Name:     browser.Name,
		Version:  browser.Version,
		Platform: browser.Platform,
		Mobile:   browser.IsMobile,
	})
	if err != nil {
		return domain.Browser{}, fmt.Errorf("failed to upsert browser: %w", err)
	}

	return domain.Browser{
		ID:       row.ID.Bytes,
		Name:     row.Name,
		Version:  row.Version,
		Platform: row.Platform,
		IsMobile: row.Mobile,
	}, nil
}

func (a urlStore) VisitOverTime(ctx context.Context, urlID domain.ID, period domain.Period, timeTrunc string) ([]domain.TimeSeriesData, error) {
	row, err := a.db.VisitOverTime(ctx, datastore.VisitOverTimeParams{
		UrlID: int32(urlID),
		StartDate: pgtype.Timestamp{
			Time:             period.Since,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		EndDate: pgtype.Timestamp{
			Time:             period.Until,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		TimeTrunc: timeTrunc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get visit per day: %w", err)
	}

	result := make([]domain.TimeSeriesData, 0, len(row))
	for _, r := range row {
		result = append(result, domain.TimeSeriesData{
			Time:  r.VisitDate.Time,
			Count: r.VisitCount,
		})
	}

	return result, nil
}
