package model

import (
	"time"

	"air-social/internal/domain"
)

type PostLike struct {
	PostID    int64     `db:"post_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

func (m *PostLike) ToDomain() *domain.PostLike {
	return &domain.PostLike{
		PostID:    m.PostID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainPostLike(d *domain.PostLike) *PostLike {
	return &PostLike{
		PostID:    d.PostID,
		UserID:    d.UserID,
		CreatedAt: d.CreatedAt,
	}
}

type CommentLike struct {
	CommentID int64     `db:"comment_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

func (m *CommentLike) ToDomain() *domain.CommentLike {
	return &domain.CommentLike{
		CommentID: m.CommentID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainCommentLike(d *domain.CommentLike) *CommentLike {
    return &CommentLike{
        CommentID: d.CommentID,
        UserID:    d.UserID,
        CreatedAt: d.CreatedAt,
    }
}