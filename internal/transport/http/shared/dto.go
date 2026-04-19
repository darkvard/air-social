package shared

// PathIDParam is a common URI path parameter for single-resource endpoints.
type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// MetaCursor holds pagination metadata for cursor-based responses.
type MetaCursor[K any] struct {
	NextCursor  K    `json:"next_cursor"`
	HasNextPage bool `json:"has_next_page"`
}

// CursorPaginatedResponse is a generic cursor-paginated response envelope.
type CursorPaginatedResponse[T any, K any] struct {
	Data []T           `json:"items"`
	Meta MetaCursor[K] `json:"meta"`
}

// UserResponse is the compact author summary embedded in post and comment responses.
type UserResponse struct {
	ID         int64  `json:"id"`
	Fullname   string `json:"full_name"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
}
