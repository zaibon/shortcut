package domain

import "time"

type URL struct {
	ID        ID
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
