package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"air-social/internal/domain"
)
// todo: remove

type Post struct {
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

func (m *Post) ToDomain() *domain.Post {
	return &domain.Post{
		ID:             m.ID,
		UserID:         m.UserID,
		Content:        m.Content,
		Visibility:     domain.PostVisibility(m.Visibility),
		Version:        m.Version,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
		OriginalPostID: m.OriginalPostID,
		Counts: domain.PostCounts{
			LikesCount:    m.LikesCount,
			CommentsCount: m.CommentsCount,
			SharesCount:   m.SharesCount,
		},
		Media: make([]domain.PostMedia, 0),
	}
}

func FromDomainPost(d *domain.Post) *Post {
	return &Post{
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

type PostMedia struct {
	ID        int64             `db:"id"`
	PostID    int64             `db:"post_id"`
	MediaKey  string            `db:"media_key"`
	MediaType string            `db:"media_type"`
	Metadata  PostMediaMetadata `db:"metadata"`
	CreatedAt time.Time         `db:"created_at"`
}

func (m *PostMedia) ToDomain() *domain.PostMedia {
	return &domain.PostMedia{
		ID:        m.ID,
		PostID:    m.PostID,
		MediaKey:  m.MediaKey,
		MediaType: m.MediaType,
		Metadata: domain.PostMediaMetadata{
			Width:    m.Metadata.Width,
			Height:   m.Metadata.Height,
			Duration: m.Metadata.Duration,
			Size:     m.Metadata.Size,
			FileName: m.Metadata.FileName,
		},
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainPostMedia(d *domain.PostMedia) *PostMedia {
	return &PostMedia{
		ID:        d.ID,
		PostID:    d.PostID,
		MediaKey:  d.MediaKey,
		MediaType: d.MediaType,
		Metadata: PostMediaMetadata{
			Width:    d.Metadata.Width,
			Height:   d.Metadata.Height,
			Duration: d.Metadata.Duration,
			Size:     d.Metadata.Size,
			FileName: d.Metadata.FileName,
		},
		CreatedAt: d.CreatedAt,
	}
}

type PostMediaMetadata struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Size     int64  `json:"size,omitempty"`
	FileName string `json:"file_name,omitempty"`
}

func (m PostMediaMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *PostMediaMetadata) Scan(value any) error {
	if value == nil {
		*m = PostMediaMetadata{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}

type PostView struct {
	Post
	IsLiked bool `db:"is_liked"`
}

func (m *PostView) ToDomain() *domain.Post {
	return &domain.Post{
		ID:             m.ID,
		UserID:         m.UserID,
		Content:        m.Content,
		Visibility:     domain.PostVisibility(m.Visibility),
		Version:        m.Version,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
		IsLiked:        m.IsLiked,
		OriginalPostID: m.OriginalPostID,
		Counts: domain.PostCounts{
			LikesCount:    m.LikesCount,
			CommentsCount: m.CommentsCount,
			SharesCount:   m.SharesCount,
		},
		Media: make([]domain.PostMedia, 0),
	}
}
