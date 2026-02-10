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
	IsFollowing  bool   `json:"is_following"`
	IsFollowedBy bool   `json:"is_followed_by"`
}
