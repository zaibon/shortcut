package handlers

import (
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid URL", "https://example.com", true},
		{"valid URL with path", "https://example.com/path", true},
		{"valid URL with query", "https://example.com?query=value", true},
		{"invalid URL missing scheme", "example.com", false},
		{"invalid URL missing host", "https://", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsValidURL(tt.url)
			if got != tt.want {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
