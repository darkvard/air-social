package search

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
