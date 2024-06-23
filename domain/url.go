package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type URL struct {
	ID        ID
	Title     string
	Long      string
	Short     string
	Slug      string
	CreatedAt time.Time
}

type URLStat struct {
	URL
	NrVisited int
	// IPAddress  []string
	// UserAgents []string
}

type URLSortRequest struct {
	SortBy  string
	SortDir string
}

func (u URLSortRequest) String() string {
	by := "created_at"
	dir := "desc"
	if strings.ToLower(u.SortDir) == "asc" {
		dir = "asc"
	}
	if slices.Contains([]string{"title", "created_at"}, strings.ToLower(u.SortBy)) {
		by = u.SortBy
	}
	return fmt.Sprintf("%s_%s", by, dir)
}
