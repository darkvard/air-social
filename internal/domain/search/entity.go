package search

import "time"

type User struct {
	ID       int64
	Username string
	Fullname string
	Avatar   string
	Verified bool
}

func (u User) GetCursor() int64 {
	return u.ID
}

// todo: media, stats...
type Post struct {
	ID        int64
	Content   string
	AuthorID  int64
	Author    User
	CreatedAt time.Time
}

func (p Post) GetCursor() int64 {
	return p.ID
}
