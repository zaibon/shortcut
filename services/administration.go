package services

import (
	"context"

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

func (s *Administration) ListUsers(ctx context.Context) ([]domain.AdminUser, error) {
	rows, err := s.db.AdminListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var users []domain.AdminUser
	for _, r := range rows {
		user := domain.AdminUser{
			User: domain.User{
				ID:        domain.ID(r.ID),
				GUID:      r.Guid.Bytes,
				Name:      r.Username,
				Email:     r.Email,
				Avatar:    "",
				CreatedAt: r.CreatedAt.Time,
				IsOauth:   r.IsOauth.Bool,
				Provider:  "",
			},
			Plan:       r.PlanName,
			URLCount:   int(r.UrlCount),
			ClickCount: int(r.ClickCount),
			Status:     r.UserStatus,
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *Administration) ListURLs(ctx context.Context) ([]domain.AdminURL, error) {
	urlsRows, err := s.db.AdminListURLSDetails(ctx)
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
				CreatedAt:  row.Url.CreatedAt.Time,
				NrVisited:  int(row.ClickCount),
			},
			Author: row.AuthorName,
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
