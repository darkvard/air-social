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

type SharerRow struct {
	ShareID    int64  `db:"share_id"`
	UserID     int64  `db:"user_id"`
	FullName   string `db:"full_name"`
	Avatar     string `db:"avatar"`
	IsVerified bool   `db:"is_verified"`
}

func (r *SharerRow) ToDomain() *post.Sharer {
	return &post.Sharer{
		ShareID:    r.ShareID,
		UserID:     r.UserID,
		FullName:   r.FullName,
		Avatar:     r.Avatar,
		IsVerified: r.IsVerified,
	}
}
