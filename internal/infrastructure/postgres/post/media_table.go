package post

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"air-social/internal/domain/post"
)

type MediaTable struct {
	ID        int64         `db:"id"`
	PostID    int64         `db:"post_id"`
	MediaKey  string        `db:"media_key"`
	MediaType string        `db:"media_type"`
	Metadata  MediaMetadata `db:"metadata"`
	CreatedAt time.Time     `db:"created_at"`
}

func (m *MediaTable) ToDomain() *post.Media {
	return &post.Media{
		ID:        m.ID,
		PostID:    m.PostID,
		MediaKey:  m.MediaKey,
		MediaType: m.MediaType,
		Metadata: post.MediaMetadata{
			Width:    m.Metadata.Width,
			Height:   m.Metadata.Height,
			Duration: m.Metadata.Duration,
			Size:     m.Metadata.Size,
			FileName: m.Metadata.FileName,
		},
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainMedia(d *post.Media) *MediaTable {
	return &MediaTable{
		ID:        d.ID,
		PostID:    d.PostID,
		MediaKey:  d.MediaKey,
		MediaType: d.MediaType,
		Metadata: MediaMetadata{
			Width:    d.Metadata.Width,
			Height:   d.Metadata.Height,
			Duration: d.Metadata.Duration,
			Size:     d.Metadata.Size,
			FileName: d.Metadata.FileName,
		},
		CreatedAt: d.CreatedAt,
	}
}

// MediaMetadata represents the JSONB structure in the database.
// It implements driver.Valuer and sql.Scanner interfaces to handle 
// automatic JSON marshaling/unmarshaling between Go structs and Postgres JSONB.
type MediaMetadata struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Size     int64  `json:"size,omitempty"`
	FileName string `json:"file_name,omitempty"`
}

// Value implements the driver.Valuer interface.
// It is called automatically by the sql driver when inserting/updating data.
// This converts the Go struct into a JSON byte array before sending it to Postgres.
func (m MediaMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface.
// It is called automatically by the sql driver when querying data (SELECT).
// This converts the JSONB byte array returned from Postgres back into the Go struct.
func (m *MediaMetadata) Scan(value any) error {
	if value == nil {
		*m = MediaMetadata{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}