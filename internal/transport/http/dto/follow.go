package dto

type FollowPathParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

