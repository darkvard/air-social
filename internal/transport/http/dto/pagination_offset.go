package dto

import "air-social/internal/domain"

type PaginationQuery struct {
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=10" binding:"min=1,max=100"`
	Sort  string `form:"sort" binding:"omitempty,oneof=latest oldest name_asc name_desc"`
}

func (q PaginationQuery) ToDomain() domain.QueryParams {
	return domain.QueryParams{
		Page:  q.Page,
		Limit: q.Limit,
		Sort:  q.Sort,
	}
}

type PaginatedResponse[T any] struct {
	Data []T        `json:"data"`
	Meta MetaPaging `json:"meta"`
}

type MetaPaging struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginatedResponse[T any](result domain.PaginatedResult[T]) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data: result.Data,
		Meta: MetaPaging{
			Total:      result.Total,
			Page:       result.Page,
			Limit:      result.Limit,
			TotalPages: result.TotalPages,
		},
	}
}
