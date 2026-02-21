package comment

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"air-social/internal/domain/comment"
)

type Media []Metadata

type Metadata struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

func (m Media) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *Media) Scan(value any) error {
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

func (m Media) ToDomain() []comment.Media {
	media := make([]comment.Media, len(m))
	for i, v := range m {
		media[i] = comment.Media{
			URL:  v.URL,
			Type: v.Type,
		}
	}
	return media
}

func FromDomainMedia(d []comment.Media) Media {
	media := make([]Metadata, len(d))
	for i, v := range d {
		media[i] = Metadata{
			URL:  v.URL,
			Type: v.Type,
		}
	}
	return media

}
