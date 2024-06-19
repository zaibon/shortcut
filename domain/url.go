package domain

type URL struct {
	ID    ID
	Long  string
	Short string
}

type URLStat struct {
	URL
	NrVisited int
	// IPAddress  []string
	// UserAgents []string
}
