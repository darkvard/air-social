package domain

// import "air-social/internal/domain/user"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

// type LoginResponse struct {
// 	User  *user.UserResponse `json:"user"`
// 	Token TokenInfo          `json:"token"`
// }

// type TokenInfo struct {
// 	AccessToken  string `json:"access_token"`
// 	RefreshToken string `json:"refresh_token,omitempty"`
// 	ExpiresIn    int64  `json:"expires_in"`
// 	ExpiresAt    string `json:"expires_at"`
// 	TokenType    string `json:"token_type,omitempty"`
// }
