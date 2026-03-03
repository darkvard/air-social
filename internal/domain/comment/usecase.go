package comment

import "context"

type UseCase interface {
	CreateComment(ctx context.Context, createParams CreateParams) (*Comment, error)

	GetRootComments(ctx context.Context, postID int64, params GetCursorParams) ([]Comment, int64, error)

	GetReplies(ctx context.Context, parentID int64, params GetCursorParams) ([]Comment, int64, error)

	UpdateComment(ctx context.Context, params UpdateParams) error

	DeleteComment(ctx context.Context, commentID, userID int64) error
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) *usecase {
	return &usecase{repo: repo}
}

func (u *usecase) CreateComment(ctx context.Context, createParams CreateParams) (*Comment, error) {
	return nil, nil
}

func (u *usecase) GetRootComments(ctx context.Context, postID int64, params GetCursorParams) ([]Comment, int64, error) {
	return nil, -1, nil
}

func (u *usecase) GetReplies(ctx context.Context, parentID int64, params GetCursorParams) ([]Comment, int64, error) {
	return nil, -1, nil
}

func (u *usecase) UpdateComment(ctx context.Context, params UpdateParams) error {
	return nil
}

func (u *usecase) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	return nil
}
