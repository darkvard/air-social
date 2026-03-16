package like

import "time"

type PostTable struct {
	ID        int64     `db:"id"`
	PostID    int64     `db:"post_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type CommentTable struct {
	ID        int64     `db:"id"`
	CommentID int64     `db:"comment_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}
