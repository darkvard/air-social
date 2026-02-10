package dto

type FollowPathParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type UserFollowResponse struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	Avatar       string `json:"avatar"`
	IsVerified   bool   `json:"is_verified"`
	IsFollowing  bool   `json:"followed_by_me"`
	IsFollowedBy bool   `json:"is_following_me"`
}
