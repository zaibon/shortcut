package domain

import "time"

type URL struct {
	ID         ID
	Title      string
	Long       string
	Short      string
	Slug       string
	IsArchived bool
	IsActive   bool
	CreatedAt  time.Time

	NrVisited int
}

type AdminURL struct {
	URL
	Author string
}

type URLStat struct {
	URL

	UniqueVisitors       int
	LocationDistribution []LocationDistribution
	Referrers            []Referrer
	ReferrersChart       []TwoDimension
	Devices              map[DeviceKind]Device
	DeviceChart          []TwoDimension
	Browsers             []BrowserStats
	BrowserChart         []TwoDimension
	VisitPerDay          []TimeSeriesData
}
type DeviceKind string

var (
	DeviceKindMobile  = DeviceKind("mobile")
	DeviceKindDesktop = DeviceKind("desktop")
)

type LocationDistribution struct {
	Country     string
	CountryCode string
	VisitCount  int
	Percentage  float32
}

type Referrer struct {
	Source     string
	ClickCount int
	Percentage float32
}

type Device struct {
	Type       string
	Percentage float32
}

type Browser struct {
	ID       GUID
	Name     string
	Version  string
	Platform string
	IsMobile bool
}

type BrowserStats struct {
	Browser    Browser
	Percentage float32
}
