package domain

import "time"

type Period struct {
	Since time.Time
	Until time.Time
}

type TimeSeriesData struct {
	Time  time.Time
	Count int64
}

type TwoDimension struct {
	Label string
	Value int
}
