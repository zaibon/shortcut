package domain

import "time"

type AdminOverview struct {
	TotalUsers  TotalCard
	TotalURLs   TotalCard
	TotalClicks TotalCard

	UserGrowth        []TimeSeriesData
	UsersOverTime     []TimeSeriesData
	URLCreationTrends []TimeSeriesData
	RecentActivity    []RecentActivity
}

type TotalCard struct {
	Total     int
	Variation int
}

type RecentActivity struct {
	Type       string
	Actor      string
	Details    string
	OccurredAt time.Time
}

// AdminUser represents a user in the admin panel with additional information
type AdminUser struct {
	User
	Plan       string
	URLCount   int
	ClickCount int
	Status     string // should be part of User
}

type AdminAnalytics struct {
	DailyUniqueVisitors []TimeSeriesData
	ClickDistribution   []TwoDimension // Referrers
	TopURLs             []TopURL
	GeoDistribution     []TwoDimension // Country + Count
}

type TopURL struct {
	ShortURL string
	LongURL  string
	Clicks   int
}

type UserFilter struct {
	Search       string
	IsSuspended  *bool
	Plan         *string
	CreatedAfter *time.Time
}

type AdminURLFilter struct {
	Search       string
	IsActive     *bool
	Plan         *string
	CreatedAfter *time.Time
	MinClicks    *int
	MaxClicks    *int
}