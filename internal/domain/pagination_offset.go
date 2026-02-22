package domain

import "math"
// todo: remove

const (
	pageMin  = 1
	limitMin = 1
	limitMax = 100
)

type QueryParams struct {
	Page  int
	Limit int
	Sort  string
}

func (q *QueryParams) EnsureDefaults() {
	if q.Page < pageMin {
		q.Page = pageMin
	}
	if q.Limit < limitMin {
		q.Limit = limitMin
	}
	if q.Limit > limitMax {
		q.Limit = limitMax
	}
}

func (q QueryParams) GetOffset() int {
	if q.Page < pageMin {
		return 0
	}
	return (q.Page - 1) * q.Limit
}

type PaginatedResult[T any] struct {
	Data       []T
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

func NewPaginatedResult[T any](data []T, total int64, page, limit int) PaginatedResult[T] {
	var totalPages int
	if limit > 0 {
		count := float64(total) / float64(limit)
		totalPages = int(math.Ceil(count))
	}
	if data == nil {
		data = []T{}
	}
	return PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
