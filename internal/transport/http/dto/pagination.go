package dto

type PaginationQuery struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=10"`
}

type PaginatedResult struct {
	Data  any   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}
