package comment

import "air-social/internal/domain/common"

type CreateParams struct {
	PostID, UserID int64
	ParentID       *int64 // Null if is the root comment
	Content        string
	Media          []Media
}

func (c CreateParams) GetMediaKeys() []string {
	if c.Media == nil {
		return []string{}
	}
	keys := make([]string, len(c.Media))
	for i, v := range c.Media {
		keys[i] = v.MediaKey
	}
	return keys
}

type UpdateParams struct {
	CommentID, UserID int64
	Content           *string
	Media             []Media
}

type GetCursorParams struct {
	ViewerID int64
	Query    common.CursorQueryParams[int64]
}
