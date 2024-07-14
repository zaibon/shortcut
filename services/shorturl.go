package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/services/geoip"
)

const idLength = 6 //TODO: make this dynamic by reading the amount of url stored in DB.

type URLStore interface {
	Add(ctx context.Context, title, shortURL, longURL string, authorID domain.ID) (domain.ID, error)
	List(ctx context.Context, authorID domain.ID, showArchived bool) ([]datastore.Url, error)
	Get(ctx context.Context, shortID string) (datastore.Url, error)
	TrackRedirect(ctx context.Context, urlID domain.ID, ipAddress, userAgent string) (datastore.Visit, error)
	InsertVisitLocation(ctx context.Context, visitID domain.ID, loc geoip.IPLocation) error
	Statistics(ctx context.Context, authorID domain.ID) ([]datastore.ListStatisticsPerAuthorRow, error)
	StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (datastore.StatisticPerURLRow, error)
	UpdateTitle(ctx context.Context, authorID domain.ID, slug, title string) (datastore.Url, error)
	ArchiveURL(ctx context.Context, authorID domain.ID, slug string) error
	UnarchiveURL(ctx context.Context, authorID domain.ID, slug string) error
}

type shortURL struct {
	repo        URLStore
	shortDomain string
}

func NewShortURL(repo URLStore, shortDomain string) *shortURL {
	return &shortURL{
		repo:        repo,
		shortDomain: shortDomain,
	}
}

// Shorten creates a new short URL, if title is empty it will try to extract it from the URL.
func (s *shortURL) Shorten(ctx context.Context, url, title string, userID domain.ID) (string, error) {
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

func (s *shortURL) ExtractTitle(url string) string {
	return ExtractTitle(url)
}

func (s *shortURL) Get(ctx context.Context, authorID domain.ID, slug string) (domain.URL, error) {
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

func (s *shortURL) List(ctx context.Context, authorID domain.ID, showArchived bool) ([]domain.URL, error) {
	rows, err := s.repo.List(ctx, authorID, showArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to list shorten urls: %w", err)
	}
	urls := make([]domain.URL, len(rows))
	for i, v := range rows {
		if v.Title == "" {
			if u, err := url.Parse(v.LongUrl); err == nil {
				v.Title = u.Host
			}
		}
		urls[i] = domain.URL{
			Title:      v.Title,
			ID:         domain.ID(v.ID),
			Long:       v.LongUrl,
			Short:      s.toURL(v.ShortUrl),
			Slug:       v.ShortUrl,
			IsArchived: v.IsArchived.Bool,
			CreatedAt:  v.CreatedAt.Time,
		}
	}

	return urls, nil
}

func (s *shortURL) Expand(ctx context.Context, short string) (domain.URL, error) {
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

func (s *shortURL) TrackRedirect(ctx context.Context, urlID domain.ID, r *http.Request) error {
	ipAddress, userAgent := parseRequest(r)
	visit, err := s.repo.TrackRedirect(ctx, urlID, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("failed to track redirect: %w", err)
	}

	if strings.Contains(ipAddress, ",") {
		ipAddress = strings.Split(ipAddress, ",")[0]
	}

	loc, err := geoip.Locate(ipAddress)
	if err != nil {
		log.Warn("failed to get country", "err", err, "ip", ipAddress)
		return nil
	}

	if err := s.repo.InsertVisitLocation(ctx, domain.ID(visit.ID), loc); err != nil {
		log.Warn("failed to insert visit location", "err", err)
	}

	return nil
}

func (s *shortURL) Statistics(ctx context.Context, authorID domain.ID) ([]domain.URLStat, error) {
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
			},
			NrVisited: int(r.NrVisits),
		}
	}

	return stats, nil
}

func (s *shortURL) StatisticsDetail(ctx context.Context, authorID domain.ID, slug string) (domain.URLStat, error) {
	detail, err := s.repo.StatisticsDetail(ctx, authorID, slug)
	if err != nil {
		return domain.URLStat{}, fmt.Errorf("failed to get statistics detail: %w", err)
	}

	return domain.URLStat{
		URL: domain.URL{
			ID:        domain.ID(detail.ID),
			Title:     detail.Title,
			Long:      detail.LongUrl,
			Short:     s.toURL(detail.ShortUrl),
			Slug:      slug,
			CreatedAt: detail.CreatedAt.Time,
		},
		NrVisited: int(detail.NrVisits),
	}, nil
}

func (s *shortURL) UpdateTitle(ctx context.Context, authorID domain.ID, slug, title string) (domain.URL, error) {
	row, err := s.repo.UpdateTitle(ctx, authorID, slug, title)
	if err != nil {
		return domain.URL{}, fmt.Errorf("failed to update title: %w", err)
	}
	return domain.URL{
		ID:        domain.ID(row.ID),
		Title:     row.Title,
		Long:      row.LongUrl,
		Short:     s.toURL(row.ShortUrl),
		Slug:      slug,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *shortURL) ArchiveURL(ctx context.Context, authorID domain.ID, slug string) error {
	return s.repo.ArchiveURL(ctx, authorID, slug)
}
func (s *shortURL) UnarchiveURL(ctx context.Context, authorID domain.ID, slug string) error {
	return s.repo.UnarchiveURL(ctx, authorID, slug)
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

func parseRequest(r *http.Request) (ipAddress string, userAgent string) {
	ipAddress = r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent = r.Header.Get("User-Agent")
	return
}

func (s *shortURL) toURL(id string) string {
	u, _ := url.JoinPath(s.shortDomain, id)
	return u
}
