package comment

import "context"

type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Comment, error)
	GetComments(ctx context.Context, postID int64, params GetCursorParams) ([]Comment, error)
	GetReplies(ctx context.Context, parentID int64, params GetCursorParams) ([]Comment, error)
}
