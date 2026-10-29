package models

const (
	DefaultPageSize = 10
	MaxPageSize     = 20
)

// Page is one page of a list, with the total number of matching rows.
type Page struct {
	Items any   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// NormalizePage applies the default page and page size, and caps the page size.
func NormalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	return page, limit
}
