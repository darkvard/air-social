package comment

import "air-social/internal/domain/comment"

type CommentRow struct {
	CommentTable
	Author
}

type Author struct {
	AuthorID int64  `db:"author_id"`
	Fullname string `db:"author_full_name"`
	Avatar   string `db:"author_avatar"`
	Verified bool   `db:"author_verified"`
}

func (r *CommentRow) ToDomain() *comment.Comment {
	c := r.CommentTable.ToDomain()
	c.Author = comment.Author{
		ID:         r.AuthorID,
		Username:   r.Fullname,
		Avatar:     r.Avatar,
		IsVerified: r.Verified,
	}
	return c
}
