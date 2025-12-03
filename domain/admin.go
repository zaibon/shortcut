package domain

type AdminOverview struct {
	TotalUsers  TotalCard
	TotalURLs   TotalCard
	TotalClicks TotalCard

	UserGrowth        []TimeSeriesData
	UsersOverTime     []TimeSeriesData
	URLCreationTrends []TimeSeriesData
}

type TotalCard struct {
	Total     int
	Variation int
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
	DailyActiveUsers  []TimeSeriesData
	ClickDistribution []TwoDimension // Referrers
	TopURLs           []TopURL
	GeoDistribution   []TwoDimension // Country + Count
}

type TopURL struct {
	ShortURL string
	LongURL  string
	Clicks   int
}