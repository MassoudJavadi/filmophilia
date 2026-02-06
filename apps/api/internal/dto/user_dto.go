package dto

// SignupRequest is what we expect from the frontend
type SignupRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,min=3"`
    Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest is what we expect for login
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// UserResponse is what we send back (excluding sensitive data like password)
type UserResponse struct {
    ID          int32  `json:"id"`
    Email       string `json:"email"`
    Username    string `json:"username"`
    DisplayName string `json:"display_name"`
}

// AuthResponse is the response for successful authentication
type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    User         UserResponse `json:"user"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}