package comment

import "air-social/internal/domain/common"

type CreateParams struct {
	PostID, UserID int64
	ParentID       *int64 // Null if is the root comment
	Content        string
	Media          []Media
}

type UpdateParams struct {
	CommentID, UserID int64
	Content           string
	Media             []Media
}

type GetCursorParams struct {
	ViewerID int64
	Query    common.CursorQueryParams[int64]
}
