package search

import "air-social/internal/domain/search"

type userRow struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Fullname string `db:"full_name"`
	Avatar   string `db:"avatar"`
	Verified bool   `db:"verified"`
}

func (r *userRow) ToDomain() search.User {
	return search.User{
		ID:       r.ID,
		Username: r.Username,
		Fullname: r.Fullname,
		Avatar:   r.Avatar,
		Verified: r.Verified,
	}
}

type postRow struct {
	// todo
}
