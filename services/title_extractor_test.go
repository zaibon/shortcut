package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestFindTitleTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty input", "", ""},
		{"no title tag", "<html><body><h1>Hello, World!</h1></body></html>", ""},
		{"valid title tag", "<html><head><title>Page Title</title></head><body><h1>Hello, World!</h1></body></html>", "Page Title"},
		// {"nested title tag", "<html><head><title><span>Nested Title</span></title></head><body><h1>Hello, World!</h1></body></html>", "Nested Title"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			z := html.NewTokenizer(strings.NewReader(tt.input))
			result := findTitleTag(z)
			assert.Equal(t, tt.expected, result)
		})
	}
}
