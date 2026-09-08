package dto

type RegisterRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=150"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	Role         string `json:"role" binding:"required,oneof=customer organizer"`
	ReferralCode string `json:"referral_code" binding:"omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}
