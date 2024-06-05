package domain

type URL struct {
	ID    int64
	Long  string
	Short string
}

type URLStat struct {
	URL
	NrVisited int
	// IPAddress  []string
	// UserAgents []string
}
