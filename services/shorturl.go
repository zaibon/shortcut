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
	Add(ctx context.Context, shortURL, longURL string, authorID domain.ID) (domain.ID, error)
	List(ctx context.Context, authorID domain.ID) ([]datastore.Url, error)
	Get(ctx context.Context, shortID string) (datastore.Url, error)
	TrackRedirect(ctx context.Context, urlID domain.ID, ipAddress, userAgent string) (datastore.Visit, error)
	InsertVisitLocation(ctx context.Context, visitID domain.ID, loc geoip.IPLocation) error
	Statistics(ctx context.Context, authorID domain.ID) ([]datastore.ListStatisticsPerAuthorRow, error)
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

func (s *shortURL) Shorten(ctx context.Context, url string, userID domain.ID) (string, error) {
	id, err := generateShortID(idLength)
	if err != nil {
		return "", err
	}

	if _, err := s.repo.Add(ctx, id, url, domain.ID(userID)); err != nil {
		return "", err
	}

	return s.toURL(id), nil
}

func (s *shortURL) List(ctx context.Context, authorID domain.ID) ([]string, error) {
	rows, err := s.repo.List(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shorten urls: %w", err)
	}
	urls := make([]string, len(rows))
	for i, v := range rows {
		urls[i] = s.toURL(v.ShortUrl)
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
			URL:       domain.URL{ID: domain.ID(r.ID), Long: r.LongUrl, Short: s.toURL(r.ShortUrl)},
			NrVisited: int(r.Visits),
		}
	}

	return stats, nil
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
