package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/services/geoip"
)

const idLength = 6 //TODO: make this dynamic by reading the amount of url stored in DB.

type URLStore interface {
	Add(ctx context.Context, title, shortURL, longURL string, authorID domain.ID) (domain.ID, error)
	List(ctx context.Context, authorID domain.ID, search string) ([]datastore.ListStatisticsPerAuthorRow, error)
	Get(ctx context.Context, slug string) (datastore.Url, error)
	GetByID(ctx context.Context, id domain.ID) (datastore.Url, error)
	Delete(ctx context.Context, urlID, authorID domain.ID) error

	TrackRedirect(ctx context.Context, urlID domain.ID, request domain.RequestInfo) (datastore.Visit, error)
	InsertVisitLocation(ctx context.Context, visitID domain.ID, loc geoip.IPLocation) error
	UpsertBrowser(ctx context.Context, requestInfo domain.Browser) (domain.Browser, error)

	Statistics(ctx context.Context, authorID domain.ID) ([]datastore.ListStatisticsPerAuthorRow, error)
	StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (datastore.StatisticPerURLRow, error)

	CountMonthlyURL(ctx context.Context, authorID domain.ID) (int64, error)
	CountMonthlyVisit(ctx context.Context, authorID domain.ID) (int64, error)
	VisitOverTime(ctx context.Context, urlID domain.ID, period domain.Period, timeTrunc string) ([]domain.TimeSeriesData, error)

	LocationDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.LocationDistributionRow, error)
	BrowserDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.BrowserDistributionRow, error)
	RefererDistribution(ctx context.Context, authorID, urlID domain.ID) ([]datastore.ReferrerDistributionRow, error)
	UniqueVisitCount(ctx context.Context, urlID domain.ID) (int64, error)
	TotalVisit(ctx context.Context, urlID domain.ID) (int64, error)
}

type urlService struct {
	repo        URLStore
	shortDomain string
}

func NewURL(repo URLStore, shortDomain string) *urlService {
	return &urlService{
		repo:        repo,
		shortDomain: shortDomain,
	}
}

// Shorten creates a new short URL, if title is empty it will try to extract it from the URL.
func (s *urlService) Shorten(ctx context.Context, url, title string, userID domain.ID) (string, error) {
	shortURL, err := generateShortID(idLength)
	if err != nil {
		return "", err
	}

	if title == "" {
		title = ExtractTitle(url)
	}

	if _, err := s.repo.Add(ctx, title, shortURL, url, domain.ID(userID)); err != nil {
		return "", err
	}

	return s.toURL(shortURL), nil
}

func (s *urlService) ExtractTitle(url string) string {
	return ExtractTitle(url)
}

func (s *urlService) Get(ctx context.Context, authorID domain.ID, slug string) (domain.URL, error) {
	row, err := s.repo.Get(ctx, slug)
	if err != nil {
		return domain.URL{}, err
	}

	return domain.URL{
		Title:      row.Title,
		ID:         domain.ID(row.ID),
		Long:       row.LongUrl,
		Short:      s.toURL(row.ShortUrl),
		Slug:       row.ShortUrl,
		IsArchived: row.IsArchived.Bool,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

func (s *urlService) GetByID(ctx context.Context, id domain.ID) (domain.URL, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.URL{}, err
	}

	return domain.URL{
		Title:      row.Title,
		ID:         domain.ID(row.ID),
		Long:       row.LongUrl,
		Short:      s.toURL(row.ShortUrl),
		Slug:       row.ShortUrl,
		IsArchived: row.IsArchived.Bool,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

func (s *urlService) List(ctx context.Context, authorID domain.ID, search string) ([]domain.URLStat, error) {
	rows, err := s.repo.List(ctx, authorID, search)
	if err != nil {
		return nil, fmt.Errorf("failed to list shorten urls: %w", err)
	}
	urls := make([]domain.URLStat, len(rows))
	for i, v := range rows {
		if v.Title == "" {
			if u, err := url.Parse(v.LongUrl); err == nil {
				v.Title = u.Host
			}
		}
		urls[i] = domain.URLStat{
			URL: domain.URL{
				Title: v.Title,
				ID:    domain.ID(v.ID),
				Long:  v.LongUrl,
				Short: s.toURL(v.ShortUrl),
				Slug:  v.ShortUrl,
				// IsArchived: v.IsArchived.Bool,
				CreatedAt: v.CreatedAt.Time,
				NrVisited: int(v.NrVisits),
			},
		}
	}

	return urls, nil
}

func (s *urlService) Delete(ctx context.Context, urlID, authorID domain.ID) error {
	return s.repo.Delete(ctx, urlID, authorID)
}

func (s *urlService) Expand(ctx context.Context, short string) (domain.URL, error) {
	item, err := s.repo.Get(ctx, short)
	if err != nil {
		return domain.URL{}, err
	}

	return domain.URL{
		ID:    domain.ID(item.ID),
		Long:  item.LongUrl,
		Short: short,
	}, nil
}

func (s *urlService) TrackRedirect(ctx context.Context, urlID domain.ID, r *http.Request) error {
	requestInfo := parseRequest(r)

	visit, err := s.repo.TrackRedirect(ctx, urlID, requestInfo)
	if err != nil {
		return fmt.Errorf("failed to track redirect: %w", err)
	}

	ip := requestInfo.IpAddress()
	loc, err := geoip.Locate(ip)
	if err != nil {
		log.Warn("failed to get country", "err", err, "ip", ip)
		return nil
	}

	if err := s.repo.InsertVisitLocation(ctx, domain.ID(visit.ID), loc); err != nil {
		log.Warn("failed to insert visit location", "err", err)
	}

	return nil
}

func (s *urlService) Statistics(ctx context.Context, authorID domain.ID) ([]domain.URLStat, error) {
	rows, err := s.repo.Statistics(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list shorten urls: %w", err)
	}
	stats := make([]domain.URLStat, len(rows))
	for i, r := range rows {
		stats[i] = domain.URLStat{
			URL: domain.URL{
				Title:     r.Title,
				ID:        domain.ID(r.ID),
				Slug:      r.ShortUrl,
				Long:      r.LongUrl,
				Short:     s.toURL(r.ShortUrl),
				CreatedAt: r.CreatedAt.Time,
				NrVisited: int(r.NrVisits),
			},
		}
	}

	return stats, nil
}

func (s *urlService) StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (domain.URLStat, error) {
	url, err := s.repo.Get(ctx, slug)
	if err != nil {
		return domain.URLStat{}, fmt.Errorf("failed to get url: %w", err)
	}
	urlID := domain.ID(url.ID)

	var (
		g, gCtx = errgroup.WithContext(ctx)
		stats   = domain.URLStat{
			URL: domain.URL{
				ID:        domain.ID(url.ID),
				Title:     url.Title,
				Long:      url.LongUrl,
				Short:     s.toURL(url.ShortUrl),
				Slug:      slug,
				CreatedAt: url.CreatedAt.Time,
				NrVisited: 0,
			},
		}
	)

	g.Go(func() error {
		ld, err := s.repo.LocationDistribution(gCtx, authorID, urlID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get location distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			ld = []datastore.LocationDistributionRow{}
		}

		for _, v := range ld {
			stats.LocationDistribution = append(stats.LocationDistribution, domain.LocationDistribution{
				Country:    v.CountryName.String,
				Percentage: float32(v.Percentage),
			})
		}

		return nil
	})

	g.Go(func() error {
		bd, err := s.repo.BrowserDistribution(gCtx, authorID, urlID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get browser distribution: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			bd = []datastore.BrowserDistributionRow{}
		}

		stats.Devices = browserDistribution(bd)
		stats.DeviceChart = devicesChart(bd)
		stats.Browsers = browsers(bd)
		stats.BrowserChart = browserChart(bd)

		return nil
	})

	g.Go(func() error {
		rd, err := s.repo.RefererDistribution(gCtx, authorID, urlID)
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
		total, err := s.repo.TotalVisit(gCtx, urlID)
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
		unique, err := s.repo.UniqueVisitCount(gCtx, urlID)
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
		now := time.Now().Truncate(time.Hour)
		since := now.AddDate(0, 0, -1).Truncate(time.Hour)
		until := now
		period := domain.Period{Since: since, Until: until}

		visitsPerDay, err := s.repo.VisitOverTime(gCtx, urlID, period, "hour")
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get unique visit count: %w", err)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			visitsPerDay = []domain.TimeSeriesData{}
		}

		stats.VisitPerDay = visitPerDay(visitsPerDay, period, time.Hour)

		return nil
	})

	if err := g.Wait(); err != nil {
		return domain.URLStat{}, err
	}

	return stats, nil
}

func (s *urlService) ClickOverTime(ctx context.Context, urlID domain.ID, period domain.Period, timeRange string) ([]domain.TimeSeriesData, error) {
	timeTrunc := "hour"
	trunc := time.Hour
	switch timeRange {
	case "day":
		timeTrunc = "hour"
		trunc = time.Hour
	case "week":
		timeTrunc = "day"
		trunc = time.Hour * 24
	case "month":
		timeTrunc = "day"
		trunc = time.Hour * 24
	}
	data, err := s.repo.VisitOverTime(ctx, urlID, period, timeTrunc)
	if err != nil {
		return nil, fmt.Errorf("failed to get visit over time: %w", err)
	}

	data = visitPerDay(data, period, trunc)

	return data, nil
}

func visitPerDay(data []domain.TimeSeriesData, period domain.Period, trunc time.Duration) []domain.TimeSeriesData {
	var (
		current = period.Since.Truncate(trunc)
		end     = period.Until.Truncate(trunc)
		newData = []domain.TimeSeriesData{}
	)

	for current.Before(end) || current.Equal(end) {
		found := false
		for _, d := range data {
			if d.Time.Equal(current) {
				newData = append(newData, d)
				found = true
				break
			}
		}
		if !found {
			newData = append(newData, domain.TimeSeriesData{
				Time:  current,
				Count: 0,
			})
		}
		current = current.Add(trunc)
	}
	return newData
}

func (s *urlService) CountMonthlyURL(ctx context.Context, authorID domain.ID) (int64, error) {
	return s.repo.CountMonthlyURL(ctx, authorID)
}

func (s *urlService) CountMonthlyVisit(ctx context.Context, authorID domain.ID) (int64, error) {
	return s.repo.CountMonthlyVisit(ctx, authorID)
}

func generateShortID(length int) (string, error) {
	// Generate a random byte slice of the desired length
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte slice to a base64 string
	s := base64.URLEncoding.EncodeToString(b)

	// Trim the string to the desired length
	return s[:length], nil
}

func parseRequest(r *http.Request) domain.RequestInfo {
	var ipAddress, userAgent, referer string
	ipAddress = r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent = r.Header.Get("User-Agent")
	referer = r.Header.Get("Referer")

	return *domain.NewRequestInfo(ipAddress, userAgent, referer)
}

func (s *urlService) toURL(id string) string {
	u, _ := url.JoinPath(s.shortDomain, id)
	return u
}

func browserDistribution(data []datastore.BrowserDistributionRow) map[domain.DeviceKind]domain.Device {
	mobile := 0
	desktop := 0
	for _, v := range data {
		if v.Mobile.Bool {
			mobile++
		} else {
			desktop++
		}
	}
	total := len(data)
	return map[domain.DeviceKind]domain.Device{
		domain.DeviceKindMobile: {
			Type:       string(domain.DeviceKindMobile),
			Percentage: float32(mobile) / float32(total) * 100,
		},
		domain.DeviceKindDesktop: {
			Type:       string(domain.DeviceKindDesktop),
			Percentage: float32(desktop) / float32(total) * 100,
		},
	}
}
func devicesChart(data []datastore.BrowserDistributionRow) []domain.TwoDimension {
	mobile := 0
	desktop := 0
	for _, v := range data {
		if v.Mobile.Bool {
			mobile++
		} else {
			desktop++
		}
	}
	return []domain.TwoDimension{
		{
			Label: "Desktop",
			Value: desktop,
		},
		{
			Label: "Mobile",
			Value: mobile,
		},
	}
}

func browsers(data []datastore.BrowserDistributionRow) []domain.BrowserStats {
	s := make([]domain.BrowserStats, len(data))
	for i, v := range data {
		s[i] = domain.BrowserStats{
			Browser: domain.Browser{
				Name:     v.Name.String,
				Version:  v.Version.String,
				Platform: v.Platform.String,
				IsMobile: v.Mobile.Bool,
			},
			Percentage: float32(v.Percentage),
		}
	}
	return s

}

func browserChart(data []datastore.BrowserDistributionRow) []domain.TwoDimension {
	m := map[string]int{}
	for _, v := range data {
		m[v.Name.String] += int(v.Percentage)
	}

	s := make([]domain.TwoDimension, 0, len(m))
	for k, v := range m {
		s = append(s, domain.TwoDimension{
			Label: k,
			Value: v,
		})
	}
	return s
}

func referers(data []datastore.ReferrerDistributionRow) []domain.Referrer {
	s := make([]domain.Referrer, len(data))
	for i, v := range data {
		s[i] = domain.Referrer{
			Source:     v.Source.String,
			ClickCount: int(v.ClickCount),
			Percentage: float32(v.Percentage),
		}
	}
	return s
}

func refererChart(data []datastore.ReferrerDistributionRow) []domain.TwoDimension {
	m := map[string]int{}
	for _, v := range data {
		m[v.Source.String] += int(v.ClickCount)
	}

	s := make([]domain.TwoDimension, 0, len(m))
	for k, v := range m {
		s = append(s, domain.TwoDimension{
			Label: k,
			Value: v,
		})
	}
	return s
}
