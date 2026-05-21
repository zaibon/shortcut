package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zaibon/shortcut/log"
)

const (
	ThreatTypeMalware           = "MALWARE"
	ThreatTypeSocialEngineering = "SOCIAL_ENGINEERING"
	ThreatTypeUnwantedSoftware  = "UNWANTED_SOFTWARE"
	ThreatTypeNone              = ""

	DefaultTimeout = 2 * time.Second
	MaxRiskScore   = 100
	MinRiskScore   = 0
)

var blockedKeywords = []string{
	"phishing",
	"malware",
	"social-engineering",
	"credential-harvesting",
	"hack-account",
	"verify-login-alert",
	"testsafebrowsing.appspot.com", // handy for local testing of heuristics
}

type SafetyScanner interface {
	Scan(ctx context.Context, targetURL string) (riskScore int, threatType string, err error)
}

type WebRiskScanner struct {
	apiKey     string
	httpClient *http.Client
}

func NewWebRiskScanner(apiKey string) *WebRiskScanner {
	return &WebRiskScanner{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (s *WebRiskScanner) Scan(ctx context.Context, targetURL string) (int, string, error) {
	// 1. Run local heuristics check first (always active)
	if isBlockedByLocalHeuristics(targetURL) {
		log.Info("Safety Scanner: URL flagged by local heuristics", "url", targetURL)
		return MaxRiskScore, ThreatTypeSocialEngineering, nil
	}

	// 2. If Google Web Risk API Key is empty, skip and treat as safe
	if s.apiKey == "" {
		return MinRiskScore, ThreatTypeNone, nil
	}

	// 3. Query Google Web Risk API
	return s.queryWebRiskAPI(ctx, targetURL)
}

func isBlockedByLocalHeuristics(targetURL string) bool {
	lowerURL := strings.ToLower(targetURL)
	for _, kw := range blockedKeywords {
		if strings.Contains(lowerURL, kw) {
			return true
		}
	}
	return false
}

type webRiskResponse struct {
	Threat struct {
		ThreatTypes []string `json:"threatTypes"`
		ExpireTime  string   `json:"expireTime"`
	} `json:"threat"`
}

func (s *WebRiskScanner) queryWebRiskAPI(ctx context.Context, targetURL string) (int, string, error) {
	apiURL := "https://webrisk.googleapis.com/v1/uris:search"

	u, err := url.Parse(apiURL)
	if err != nil {
		return MinRiskScore, ThreatTypeNone, fmt.Errorf("failed to parse Web Risk API URL: %w", err)
	}

	q := u.Query()
	q.Set("key", s.apiKey)
	q.Set("uri", targetURL)
	q.Add("threatTypes", ThreatTypeMalware)
	q.Add("threatTypes", ThreatTypeSocialEngineering)
	q.Add("threatTypes", ThreatTypeUnwantedSoftware)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return MinRiskScore, ThreatTypeNone, fmt.Errorf("failed to create Web Risk request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Log the error but fail-safe so user flow is not entirely broken if Google Web Risk is down
		log.Error("Safety Scanner: Google Web Risk API query error", "err", err)
		return MinRiskScore, ThreatTypeNone, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Info("Safety Scanner: Google Web Risk API returned non-OK status", "status", resp.Status)
		return MinRiskScore, ThreatTypeNone, nil
	}

	var data webRiskResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return MinRiskScore, ThreatTypeNone, fmt.Errorf("failed to decode Web Risk response: %w", err)
	}

	if len(data.Threat.ThreatTypes) > 0 {
		threat := data.Threat.ThreatTypes[0]
		log.Info("Safety Scanner: Google Web Risk API flagged URL", "url", targetURL, "threat", threat)
		return MaxRiskScore, threat, nil
	}

	return MinRiskScore, ThreatTypeNone, nil
}
