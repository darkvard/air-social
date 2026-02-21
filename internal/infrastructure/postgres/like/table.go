package like

import "time"

type PostTable struct {
	PostID    int64     `db:"post_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type CommentTable struct {
	CommentID int64     `db:"comment_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}
