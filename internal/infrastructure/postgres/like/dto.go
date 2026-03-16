package like

import "air-social/internal/domain/like"

type LikerRow struct {
	LikeID     int64  `db:"like_id"`
	UserID     int64  `db:"user_id"`
	FullName   string `db:"full_name"`
	Avatar     string `db:"avatar"`
	IsVerified bool   `db:"is_verified"`
}

func (r *LikerRow) ToDomain() *like.Liker {
	return &like.Liker{
		LikeID:     r.LikeID,
		UserID:     r.UserID,
		FullName:   r.FullName,
		Avatar:     r.Avatar,
		IsVerified: r.IsVerified,
	}
}
