package domain

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type LoginRequest struct {
	Email    string `json:"email,omitempty" binding:"required,email,max=255"`
	Password string `json:"password,omitempty" binding:"required,min=8,max=64"`
}
