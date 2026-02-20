package post

import (
	"time"

	"air-social/internal/domain/post"
)

type PostTable struct {
	ID             int64      `db:"id"`
	UserID         int64      `db:"user_id"`
	Content        string     `db:"content"`
	Visibility     string     `db:"visibility"`
	Version        int        `db:"version"`
	LikesCount     int        `db:"likes_count"`
	CommentsCount  int        `db:"comments_count"`
	SharesCount    int        `db:"shares_count"`
	OriginalPostID *int64     `db:"original_post_id"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
}

func (m *PostTable) ToDomain() *post.Post {
	return &post.Post{
		ID:             m.ID,
		UserID:         m.UserID,
		Content:        m.Content,
		Visibility:     post.Visibility(m.Visibility),
		Version:        m.Version,
		OriginalPostID: m.OriginalPostID,
		Counts: post.Counts{
			LikesCount:    m.LikesCount,
			CommentsCount: m.CommentsCount,
			SharesCount:   m.SharesCount,
		},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Media:     []post.Media{},
	}
}

func FromDomainPost(d *post.Post) *PostTable {
	return &PostTable{
		ID:             d.ID,
		UserID:         d.UserID,
		Content:        d.Content,
		Visibility:     string(d.Visibility),
		Version:        d.Version,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
		OriginalPostID: d.OriginalPostID,
		LikesCount:     d.Counts.LikesCount,
		CommentsCount:  d.Counts.CommentsCount,
		SharesCount:    d.Counts.SharesCount,
	}
}
