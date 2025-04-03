package domain

import "time"

type URL struct {
	ID         ID
	Title      string
	Long       string
	Short      string
	Slug       string
	IsArchived bool
	CreatedAt  time.Time

	NrVisited int
}

type URLStat struct {
	URL

	UniqueVisitors       int
	LocationDistribution []LocationDistribution
	Referrers            []Referrer
	Devices              map[DeviceKind]Device
	Browsers             []Browser
}
type DeviceKind string

var (
	DeviceKindMobile  = DeviceKind("mobile")
	DeviceKindDesktop = DeviceKind("desktop")
)

type LocationDistribution struct {
	Country    string
	Percentage float32
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
	Name       string
	Version    string
	Platform   string
	Percentage float32
}
