package post

import "air-social/internal/domain/shared"

type CreateParams struct {
	UserID     int64
	Content    string
	Visibility Visibility
	Media      []MediaParams
}

type MediaParams struct {
	MediaKey  string
	MediaType string
	Width     int
	Height    int
	Duration  int
	Size      int64
	FileName  string
}

type UpdateParams struct {
	UserID     int64
	PostID     int64
	Content    *string
	Visibility *Visibility
}

type GetCursorParams struct {
	UserID int64
	Query  shared.CursorQueryParams[int64]
}

func (p MediaParams) ToDomain() Media {
	return Media{
		MediaKey:  p.MediaKey,
		MediaType: p.MediaType,
		Metadata: MediaMetadata{
			Width:    p.Width,
			Height:   p.Height,
			Duration: p.Duration,
			Size:     p.Size,
			FileName: p.FileName,
		},
	}
}
