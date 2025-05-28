package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zaibon/shortcut/domain"
)

// ContextKey is the context key type for pagination params.
type ContextKey string

// PaginationParamsKey is the context key for pagination params.
const PaginationParamsKey ContextKey = "paginationParams"

// PaginationParams holds pagination parameters.
type PaginationParams struct {
	Page         int
	PageSize     int
	TotalRecords int
}

func (q *PaginationParams) Offset() int {
	perPage := q.PageSize
	if perPage <= 0 {
		perPage = 10
	}
	offset := (q.Page - 1) * perPage
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (q *PaginationParams) Limit() int {
	limit := q.PageSize

	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return limit
}

func Paginate[T any](s []T, q PaginationParams) []T {
	start := q.Offset()
	end := q.Offset() + q.Limit()
	switch {
	case start >= len(s):
		return s[:0]
	case end > len(s):
		return s[start:]
	default:
		return s[start:end]
	}
}

// PaginateParams middleware
func PaginateParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default values
		page := 1
		pageSize := 10

		// Get query parameters
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("page_size")

		// Parse page
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err == nil && p > 0 {
				page = p
			}
		}

		// Parse page size
		if pageSizeStr != "" {
			ps, err := strconv.Atoi(pageSizeStr)
			if err == nil && ps > 0 {
				pageSize = ps
			}
		}

		// Create PaginationParams
		params := PaginationParams{
			Page:     page,
			PageSize: pageSize,
		}

		// Put params to context
		ctx := context.WithValue(r.Context(), PaginationParamsKey, params)

		// Call the next handler, which can read from the context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPaginationParams retrieves pagination parameters from the context.
func GetPaginationParams(ctx context.Context) PaginationParams {
	params, ok := ctx.Value(PaginationParamsKey).(PaginationParams)
	if !ok {
		return PaginationParams{Page: 1, PageSize: 10} // Default values if not found
	}
	return params
}

// SetTotalRecords sets the total number of records in the PaginationParams stored in the context.
func SetTotalRecords(ctx context.Context, totalRecords int) context.Context {
	params := GetPaginationParams(ctx)
	params.TotalRecords = totalRecords
	return context.WithValue(ctx, PaginationParamsKey, params)
}

// func ExampleHandler(w http.ResponseWriter, r *http.Request) {
// 	// Get pagination parameters from the context
// 	params := GetPaginationParams(r.Context())

// 	// Example usage:
// 	// - Query the database with the pagination parameters
// 	// - Set the total number of records using SetTotalRecords
// 	// - Return the paginated data in the response

// 	w.Write([]byte("Page: " + strconv.Itoa(params.Page) + ", PageSize: " + strconv.Itoa(params.PageSize)))
// }

// func main() {
// 	r := chi.NewRouter()
// 	r.Use(Paginate)
// 	r.Get("/", ExampleHandler)
// 	http.ListenAndServe(":3000", r)
// }

// GeneratePaginationLinks generates a list of pagination links.
func GeneratePaginationLinks(params PaginationParams, total int) domain.PaginationLinks {
	var links []domain.Link
	totalPages := (total + params.PageSize - 1) / params.PageSize

	// Calculate the range of pages to display
	startPage := max(1, params.Page-2)
	endPage := min(totalPages, params.Page+2)

	// Adjust start and end pages if the range is too small at the beginning or end
	if endPage-startPage+1 < 5 {
		if startPage == 1 {
			endPage = min(totalPages, startPage+4)
		} else if endPage == totalPages {
			startPage = max(1, endPage-4)
		}
	}

	// Previous page link
	var previousLink *domain.Link
	if params.Page > 1 {
		previousLink = &domain.Link{
			Label: "previous",
			Href:  fmt.Sprintf("?page=%d&page_size=%d", params.Page-1, params.PageSize),
		}
	}

	// Page links
	for i := startPage; i <= endPage; i++ {
		links = append(links, domain.Link{
			Href:    fmt.Sprintf("?page=%d&page_size=%d", i, params.PageSize),
			Label:   fmt.Sprintf("%d", i),
			Current: i == params.Page,
		})
	}

	// Next page link
	var nextLink *domain.Link
	if params.Page < totalPages {
		nextLink = &domain.Link{
			Label: "next",
			Href:  fmt.Sprintf("?page=%d&page_size=%d", params.Page+1, params.PageSize),
		}
	}

	return domain.PaginationLinks{
		Previous: previousLink,
		Pages:    links,
		Next:     nextLink,
	}
}
