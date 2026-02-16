package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"air-social/internal/domain"
)

type Comment struct {
	ID         int64        `db:"id"`
	PostID     int64        `db:"post_id"`
	UserID     int64        `db:"user_id"`
	ParentID   *int64       `db:"parent_id"`
	Content    string       `db:"content"`
	Media      CommentMedia `db:"media"`
	LikesCount int          `db:"likes_count"`
	Version    int          `db:"version"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	DeletedAt  *time.Time   `db:"deleted_at"`
}

func (m *Comment) ToDomain() *domain.Comment {
	media := make([]domain.CommentMedia, len(m.Media))
	for i, v := range m.Media {
		media[i] = domain.CommentMedia{
			URL:  v.URL,
			Type: v.Type,
		}
	}

	return &domain.Comment{
		ID:         m.ID,
		PostID:     m.PostID,
		UserID:     m.UserID,
		ParentID:   m.ParentID,
		Content:    m.Content,
		Media:      media,
		LikesCount: m.LikesCount,
		Version:    m.Version,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  m.DeletedAt,
	}
}

func FromDomainComment(d *domain.Comment) *Comment {
	media := make(CommentMedia, len(d.Media))
	for i, item := range d.Media {
		media[i] = CommentMediaItem{
			URL:  item.URL,
			Type: item.Type,
		}
	}

	return &Comment{
		ID:         d.ID,
		PostID:     d.PostID,
		UserID:     d.UserID,
		ParentID:   d.ParentID,
		Content:    d.Content,
		Media:      media,
		LikesCount: d.LikesCount,
		Version:    d.Version,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		DeletedAt:  d.DeletedAt,
	}
}

type CommentMedia []CommentMediaItem

type CommentMediaItem struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

func (m CommentMedia) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *CommentMedia) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(b, m)
}

type CommentView struct {
	Comment
	IsLiked bool `db:"is_liked"`
}

func (m *CommentView) ToDomain() *domain.Comment {
	media := make([]domain.CommentMedia, len(m.Media))
	for i, v := range m.Media {
		media[i] = domain.CommentMedia{
			URL:  v.URL,
			Type: v.Type,
		}
	}

	return &domain.Comment{
		ID:         m.ID,
		PostID:     m.PostID,
		UserID:     m.UserID,
		ParentID:   m.ParentID,
		Content:    m.Content,
		Media:      media,
		LikesCount: m.LikesCount,
		Version:    m.Version,
		IsLiked:    m.IsLiked,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  m.DeletedAt,
	}
}
