package domain

type Link struct {
	Href    string
	Label   string
	Current bool
}

type PaginationLinks struct {
	Previous *Link
	Pages    []Link
	Next     *Link

	Min        int
	Max        int
	TotalItems int
}
