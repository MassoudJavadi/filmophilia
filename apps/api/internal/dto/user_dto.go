package dto

// SignupRequest is what we expect from the frontend
type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

// UserResponse is what we send back (excluding sensitive data like password)
type UserResponse struct {
	ID          int32  `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}