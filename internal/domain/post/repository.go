package post

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, postID int64) (*Post, error)
	GetDetail(ctx context.Context, postID int64) (*Post, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*Post, error)
	GetUserPosts(ctx context.Context, params GetCursorParams) ([]Post, error)
	GetPostSharers(ctx context.Context, params GetSharersParams) ([]Sharer, error)
}
