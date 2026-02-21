package post

import "air-social/internal/domain/post"

type PostDetailRow struct {
	PostTable
	Author
}

type Author struct {
	AuthorID int64  `db:"author_id"`
	Fullname string `db:"author_full_name"`
	Avatar   string `db:"author_avatar"`
	Verified bool   `db:"author_verified"`
}

func (m *PostDetailRow) ToDomain() *post.Post {
	domain := m.PostTable.ToDomain()
	domain.Author = &post.Author{
		ID:         m.AuthorID,
		FullName:   m.Fullname,
		Avatar:     m.Avatar,
		IsVerified: m.Verified,
	}
	return domain
}
