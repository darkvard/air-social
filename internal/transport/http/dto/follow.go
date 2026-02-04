package dto

type FollowRequest struct {
	TargetUserID int64 `uri:"id" binding:"required,gt=0"`
}
