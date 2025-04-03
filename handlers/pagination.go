package handlers

type PaginationQuery struct {
	Page    int `url:"page"`
	PerPage int `url:"perPage"`
}

type PaginateResponse struct {
	TotalItems int `url:"totalItems" json:"totalItems"`
	PerPage    int `url:"perPage" json:"perPage"`
	Page       int `url:"page" json:"page"`
}

func (q *PaginationQuery) Offset() int {
	perPage := q.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	offset := (q.Page - 1) * perPage
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (q *PaginationQuery) Limit() int {
	limit := q.PerPage

	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	return limit
}
