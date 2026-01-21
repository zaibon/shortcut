package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
)

type Administration struct {
	db     datastore.Querier
	domain string
}

func NewAdministrationService(db datastore.DBTX, domain string) *Administration {
	return &Administration{
		db:     datastore.New(db),
		domain: domain,
	}
}

func (s *Administration) GetUser(ctx context.Context, guid domain.GUID) (domain.AdminUser, error) {
	r, err := s.db.AdminGetUser(ctx, pgtype.UUID{Bytes: guid, Valid: true})
	if err != nil {
		return domain.AdminUser{}, err
	}

	return domain.AdminUser{
		User: domain.User{
			ID:          domain.ID(r.ID),
			GUID:        r.Guid.Bytes,
			Name:        r.Username,
			Email:       r.Email,
			Avatar:      "",
			CreatedAt:   r.CreatedAt.Time,
			IsOauth:     r.IsOauth.Bool,
			Provider:    "",
			IsSuspended: r.UserStatus == "suspended",
		},
		Plan:       r.PlanName,
		URLCount:   int(r.UrlCount),
		ClickCount: int(r.ClickCount),
		Status:     r.UserStatus,
	}, nil
}

func (s *Administration) ListUsers(ctx context.Context, filter domain.UserFilter) ([]domain.AdminUser, error) {

	params := datastore.AdminListUsersParams{

		Search: pgtype.Text{

			String: filter.Search,

			Valid: filter.Search != "",
		},
	}

	if filter.IsSuspended != nil {

		params.IsSuspended = pgtype.Bool{

			Bool: *filter.IsSuspended,

			Valid: true,
		}

	}

	if filter.Plan != nil {

		params.Plan = pgtype.Text{

			String: *filter.Plan,

			Valid: true,
		}

	}

	if filter.CreatedAfter != nil {

		params.CreatedAfter = pgtype.Timestamp{

			Time: *filter.CreatedAfter,

			Valid: true,
		}

	}

	rows, err := s.db.AdminListUsers(ctx, params)

	if err != nil {

		return nil, err

	}

	var users []domain.AdminUser

	for _, r := range rows {

		user := domain.AdminUser{

			User: domain.User{

				ID: domain.ID(r.ID),

				GUID: r.Guid.Bytes,

				Name: r.Username,

				Email: r.Email,

				Avatar: "",

				CreatedAt: r.CreatedAt.Time,

				IsOauth: r.IsOauth.Bool,

				Provider: "",

				IsSuspended: r.UserStatus == "suspended",
			},

			Plan: r.PlanName,

			URLCount: int(r.UrlCount),

			ClickCount: int(r.ClickCount),

			Status: r.UserStatus,
		}

		users = append(users, user)

	}

	return users, nil

}

func (s *Administration) ListURLs(ctx context.Context, filter domain.AdminURLFilter) ([]domain.AdminURL, error) {
	params := datastore.AdminListURLSDetailsParams{
		Search: pgtype.Text{
			String: filter.Search,
			Valid:  filter.Search != "",
		},
	}

	if filter.IsActive != nil {
		params.IsActive = pgtype.Bool{
			Bool:  *filter.IsActive,
			Valid: true,
		}
	}

	if filter.Plan != nil {
		params.Plan = pgtype.Text{
			String: *filter.Plan,
			Valid:  true,
		}
	}

	if filter.CreatedAfter != nil {
		params.CreatedAfter = pgtype.Timestamp{
			Time:  *filter.CreatedAfter,
			Valid: true,
		}
	}

	if filter.MinClicks != nil {
		params.MinClicks = pgtype.Int4{
			Int32: int32(*filter.MinClicks),
			Valid: true,
		}
	}

	if filter.MaxClicks != nil {
		params.MaxClicks = pgtype.Int4{
			Int32: int32(*filter.MaxClicks),
			Valid: true,
		}
	}

	urlsRows, err := s.db.AdminListURLSDetails(ctx, params)
	if err != nil {
		return nil, err
	}

	var urls []domain.AdminURL
	for _, row := range urlsRows {
		url := domain.AdminURL{
			URL: domain.URL{
				ID:         domain.ID(row.Url.ID),
				Title:      row.Url.Title,
				Long:       row.Url.LongUrl,
				Short:      toURL(s.domain, row.Url.ShortUrl),
				Slug:       row.Url.ShortUrl,
				IsArchived: row.Url.IsArchived.Bool,
				IsActive:   row.Url.IsActive,
				CreatedAt:  row.Url.CreatedAt.Time,
				NrVisited:  int(row.ClickCount),
			},
			Author:     row.AuthorName,
			AuthorGUID: domain.GUID(row.User.Guid.Bytes),
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (s *Administration) GetUserURLs(ctx context.Context, guid domain.GUID) ([]domain.AdminURL, error) {
	urlsRows, err := s.db.AdminListUserURLs(ctx, pgtype.UUID{Bytes: guid, Valid: true})
	if err != nil {
		return nil, err
	}

	var urls []domain.AdminURL
	for _, row := range urlsRows {
		url := domain.AdminURL{
			URL: domain.URL{
				ID:         domain.ID(row.Url.ID),
				Title:      row.Url.Title,
				Long:       row.Url.LongUrl,
				Short:      toURL(s.domain, row.Url.ShortUrl),
				Slug:       row.Url.ShortUrl,
				IsArchived: row.Url.IsArchived.Bool,
				IsActive:   row.Url.IsActive,
				CreatedAt:  row.Url.CreatedAt.Time,
				NrVisited:  int(row.ClickCount),
			},
			Author:     row.AuthorName,
			AuthorGUID: domain.GUID(row.User.Guid.Bytes),
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (s *Administration) GetOverviewStats(ctx context.Context) (*domain.AdminOverview, error) {
	stats, err := s.db.AdminGetOverviewStatistics(ctx)
	if err != nil {
		return nil, err
	}

	userGrowsRows, err := s.db.AdminGetUserGrowth(ctx)
	if err != nil {
		return nil, err
	}

	var userGrows []domain.TimeSeriesData
	for _, row := range userGrowsRows {
		userGrows = append(userGrows, domain.TimeSeriesData{
			Time:  row.Day.Time,
			Count: row.Count,
		})
	}

	urlTrends, err := s.db.AdminGetURLCreationTrends(ctx)
	if err != nil {
		return nil, err
	}
	var urlCreationTrends []domain.TimeSeriesData
	for _, row := range urlTrends {
		urlCreationTrends = append(urlCreationTrends, domain.TimeSeriesData{
			Time:  row.Day.Time,
			Count: row.Count,
		})
	}

	usersTotalRows, err := s.db.AdminGetTotalUsersTrend(ctx)
	if err != nil {
		return nil, err
	}
	var usersTotal []domain.TimeSeriesData
	for _, row := range usersTotalRows {
		usersTotal = append(usersTotal, domain.TimeSeriesData{
			Time:  row.Day.Time,
			Count: row.Count,
		})
	}

	overview := &domain.AdminOverview{
		TotalUsers: domain.TotalCard{
			Total:     int(stats.TotalUsers),
			Variation: int(stats.TotalUsersVariation.(float64)),
		},
		TotalURLs: domain.TotalCard{
			Total:     int(stats.TotalUrls),
			Variation: int(stats.TotalUrlsVariation.(float64)),
		},
		TotalClicks: domain.TotalCard{
			Total:     int(stats.TotalClicks),
			Variation: int(stats.TotalClicksVariation.(float64)),
		},
		UserGrowth:        userGrows,
		UsersOverTime:     usersTotal,
		URLCreationTrends: urlCreationTrends,
	}

	return overview, nil
}

func (s *Administration) GetAnalyticsStats(ctx context.Context) (*domain.AdminAnalytics, error) {
	dailyActive, err := s.db.AdminGetDailyActiveVisitors(ctx)
	if err != nil {
		return nil, err
	}
	var dailyActiveSeries []domain.TimeSeriesData
	for _, da := range dailyActive {
		dailyActiveSeries = append(dailyActiveSeries, domain.TimeSeriesData{
			Time:  da.Day.Time,
			Count: da.Count,
		})
	}

	topReferrers, err := s.db.AdminGetTopReferrers(ctx)
	if err != nil {
		return nil, err
	}
	var referrers []domain.TwoDimension
	for _, ref := range topReferrers {
		source, ok := ref.Source.(string)
		if !ok {
			source = "Unknown"
		}
		referrers = append(referrers, domain.TwoDimension{
			Label: source,
			Value: int(ref.Count),
		})
	}

	topURLsRows, err := s.db.AdminGetTopURLs(ctx)
	if err != nil {
		return nil, err
	}
	var topURLs []domain.TopURL
	for _, u := range topURLsRows {
		topURLs = append(topURLs, domain.TopURL{
			ShortURL: toURL(s.domain, u.ShortUrl),
			LongURL:  u.LongUrl,
			Clicks:   int(u.Clicks),
		})
	}

	geoRows, err := s.db.AdminGetGeoDistribution(ctx)
	if err != nil {
		return nil, err
	}
	var geoDist []domain.TwoDimension
	for _, g := range geoRows {
		geoDist = append(geoDist, domain.TwoDimension{
			Label: g.Country,
			Value: int(g.Count),
		})
	}

	return &domain.AdminAnalytics{
		DailyActiveUsers:  dailyActiveSeries,
		ClickDistribution: referrers,
		TopURLs:           topURLs,
		GeoDistribution:   geoDist,
	}, nil
}

func (s *Administration) DeleteURL(ctx context.Context, id domain.ID) error {
	return s.db.AdminDeleteURL(ctx, int32(id))
}

func (s *Administration) ToggleURLStatus(ctx context.Context, id domain.ID, isArchived, isActive bool) error {
	return s.db.AdminUpdateURLStatus(ctx, datastore.AdminUpdateURLStatusParams{
		ID:         int32(id),
		IsArchived: pgtype.Bool{Bool: isArchived, Valid: true},
		IsActive:   isActive,
	})
}

func (s *Administration) ToggleUserSuspension(ctx context.Context, guid domain.GUID, isSuspended bool) error {
	return s.db.UpdateUserSuspension(ctx, datastore.UpdateUserSuspensionParams{
		Guid: pgtype.UUID{
			Bytes: guid,
			Valid: true,
		},
		IsSuspended: isSuspended,
	})
}

func (s *Administration) ToggleUserURLsStatus(ctx context.Context, guid domain.GUID, isActive bool) error {
	return s.db.AdminToggleUserURLs(ctx, datastore.AdminToggleUserURLsParams{
		Guid: pgtype.UUID{
			Bytes: guid,
			Valid: true,
		},
		IsActive: isActive,
	})
}

func (s *Administration) UpdateURL(ctx context.Context, id domain.ID, title, longURL string) error {
	_, err := s.db.AdminUpdateURL(ctx, datastore.AdminUpdateURLParams{
		ID:      int32(id),
		Title:   title,
		LongUrl: longURL,
	})
	return err
}

func (s *Administration) GetURL(ctx context.Context, id domain.ID) (domain.URL, error) {
	row, err := s.db.GetByID(ctx, int32(id))
	if err != nil {
		return domain.URL{}, err
	}
	return domain.URL{
		ID:         domain.ID(row.ID),
		Title:      row.Title,
		Long:       row.LongUrl,
		Short:      toURL(s.domain, row.ShortUrl),
		Slug:       row.ShortUrl,
		IsArchived: row.IsArchived.Bool,
		IsActive:   row.IsActive,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

func (s *Administration) GetURLStats(ctx context.Context, slug string) (domain.URLStat, error) {
	url, err := s.db.GetShortURL(ctx, slug)
	if err != nil {
		return domain.URLStat{}, fmt.Errorf("failed to get url: %w", err)
	}
	urlID := domain.ID(url.ID)
	authorID := domain.ID(url.AuthorID)

	var (
		g, gCtx = errgroup.WithContext(ctx)
		stats   = domain.URLStat{
			URL: domain.URL{
				ID:         domain.ID(url.ID),
				Title:      url.Title,
				Long:       url.LongUrl,
				Short:      toURL(s.domain, url.ShortUrl),
				Slug:       slug,
				IsArchived: url.IsArchived.Bool,
				IsActive:   url.IsActive,
				CreatedAt:  url.CreatedAt.Time,
				NrVisited:  0,
			},
		}
	)

	g.Go(func() error {
		ld, err := s.db.LocationDistribution(gCtx, datastore.LocationDistributionParams{
			UrlID:    int32(urlID),
			AuthorID: int32(authorID),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get location distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			ld = []datastore.LocationDistributionRow{}
		}

		for _, v := range ld {
			stats.LocationDistribution = append(stats.LocationDistribution, domain.LocationDistribution{
				Country:     v.CountryName.String,
				CountryCode: v.CountryCode.String,
				VisitCount:  int(v.VisitCount),
				Percentage:  float32(v.Percentage),
			})
		}

		return nil
	})

	g.Go(func() error {
		bd, err := s.db.BrowserDistribution(gCtx, datastore.BrowserDistributionParams{
			UrlID:    int32(urlID),
			AuthorID: int32(authorID),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get browser distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			bd = []datastore.BrowserDistributionRow{}
		}

		stats.Browsers = browsers(bd)
		stats.BrowserChart = browserChart(bd)

		return nil
	})

	g.Go(func() error {
		dd, err := s.db.DeviceDistribution(gCtx, datastore.DeviceDistributionParams{
			UrlID:    int32(urlID),
			AuthorID: int32(authorID),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get device distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			dd = []datastore.DeviceDistributionRow{}
		}

		stats.Devices = calculateDeviceStats(dd)
		stats.DeviceChart = devicesChart(dd)

		return nil
	})

	g.Go(func() error {
		rd, err := s.db.ReferrerDistribution(gCtx, datastore.ReferrerDistributionParams{
			UrlID:    int32(urlID),
			AuthorID: int32(authorID),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get referer distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			rd = []datastore.ReferrerDistributionRow{}
		}

		stats.Referrers = referers(rd)
		stats.ReferrersChart = refererChart(rd)

		return nil
	})

	g.Go(func() error {
		total, err := s.db.TotalVisit(gCtx, int32(urlID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get total visit: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			total = 0
		}
		stats.NrVisited = int(total)
		return nil
	})

	g.Go(func() error {
		unique, err := s.db.UniqueVisitCount(gCtx, int32(urlID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get unique visit count: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			unique = 0
		}

		stats.UniqueVisitors = int(unique)
		return nil
	})

	g.Go(func() error {
		now := time.Now()
		since := now.AddDate(0, 0, -1).Truncate(time.Hour)
		until := now
		period := domain.Period{Since: since, Until: until}

		visitsPerDay, err := s.db.VisitOverTime(gCtx, datastore.VisitOverTimeParams{
			TimeTrunc: "hour",
			StartDate: pgtype.Timestamp{Time: since, Valid: true},
			EndDate:   pgtype.Timestamp{Time: until, Valid: true},
			UrlID:     int32(urlID),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get unique visit count: %w", err)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			visitsPerDay = []datastore.VisitOverTimeRow{}
		}

		var domainVisits []domain.TimeSeriesData
		for _, v := range visitsPerDay {
			domainVisits = append(domainVisits, domain.TimeSeriesData{
				Time:  v.VisitDate.Time,
				Count: v.VisitCount,
			})
		}
		stats.VisitPerDay = visitPerDay(domainVisits, period, time.Hour)

		return nil
	})

	if err := g.Wait(); err != nil {
		return domain.URLStat{}, err
	}

	return stats, nil
}
