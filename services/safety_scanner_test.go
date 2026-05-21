package services

import (
	"context"
	"testing"
)

func TestSafetyScanner_LocalHeuristics(t *testing.T) {
	tests := []struct {
		name       string
		targetURL  string
		wantScore  int
		wantThreat string
	}{
		{
			name:       "flagged by phishing keyword",
			targetURL:  "https://myphishingwebsite.com/login",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
		{
			name:       "flagged by malware keyword",
			targetURL:  "http://example.org/download-malware-now",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
		{
			name:       "flagged by testsafebrowsing domain",
			targetURL:  "https://testsafebrowsing.appspot.com/s/phishing.html",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
		{
			name:       "flagged by social engineering",
			targetURL:  "http://evil-social-engineering.com",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
		{
			name:       "flagged by credential harvesting",
			targetURL:  "https://credential-harvesting-alert.net",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
		{
			name:       "flagged by verify login alert",
			targetURL:  "https://verify-login-alert.info/secure",
			wantScore:  MaxRiskScore,
			wantThreat: ThreatTypeSocialEngineering,
		},
	}

	// Scanner with no API key should still flag heuristics
	scanner := NewWebRiskScanner("")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, threat, err := scanner.Scan(context.Background(), tt.targetURL)
			if err != nil {
				t.Fatalf("unexpected error during scan: %v", err)
			}
			if score != tt.wantScore {
				t.Errorf("Scan() score = %v, want %v", score, tt.wantScore)
			}
			if threat != tt.wantThreat {
				t.Errorf("Scan() threat = %v, want %v", threat, tt.wantThreat)
			}
		})
	}
}

func TestSafetyScanner_CleanURLNoAPIKey(t *testing.T) {
	scanner := NewWebRiskScanner("")

	cleanURLs := []string{
		"https://google.com",
		"https://github.com/zaibon/shortcut",
		"https://pkg.go.dev/net/http",
	}

	for _, u := range cleanURLs {
		t.Run(u, func(t *testing.T) {
			score, threat, err := scanner.Scan(context.Background(), u)
			if err != nil {
				t.Fatalf("unexpected error scanning clean URL: %v", err)
			}
			if score != MinRiskScore {
				t.Errorf("Scan() score = %v, want %v for clean URL", score, MinRiskScore)
			}
			if threat != ThreatTypeNone {
				t.Errorf("Scan() threat = %q, want %q for clean URL", threat, ThreatTypeNone)
			}
		})
	}
}
