package comment

import (
	"time"

	"air-social/internal/domain/comment"
)

type CommentTable struct {
	ID        int64      `db:"id"`
	PostID    int64      `db:"post_id"`
	UserID    int64      `db:"user_id"`
	ParentID  *int64     `db:"parent_id"`
	Content   string     `db:"content"`
	Media     Media      `db:"media"`
	Version   int32      `db:"version"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (m *CommentTable) ToDomain() *comment.Comment {
	return &comment.Comment{
		ID:        m.ID,
		PostID:    m.PostID,
		UserID:    m.UserID,
		ParentID:  m.ParentID,
		Content:   m.Content,
		Media:     m.Media.ToDomain(),
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func FromDomainComment(d *comment.Comment) *CommentTable {
	if d == nil {
		return nil
	}
	return &CommentTable{
		ID:        d.ID,
		PostID:    d.PostID,
		UserID:    d.UserID,
		ParentID:  d.ParentID,
		Content:   d.Content,
		Media:     FromDomainMedia(d.Media),
		Version:   d.Version,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}
