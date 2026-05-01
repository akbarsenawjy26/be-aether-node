package auth

import "context"

type LoginRequest struct {
	Email    string
	Password string
}

type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type ForgotPasswordRequest struct {
	Email string
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type AuthService interface {
	// Login authenticates a user and returns tokens
	Login(ctx context.Context, req *LoginRequest) (*LoginResult, error)

	// Logout invalidates the user's refresh token
	Logout(ctx context.Context, userGUID string) error

	// Register creates a new user account
	Register(ctx context.Context, req *RegisterRequest) (*UserInfo, error)

	// ForgotPassword sends a password reset email
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error

	// RefreshToken exchanges a refresh token for new tokens
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResult, error)

	// ValidateAccessToken validates an access token and returns the user GUID
	ValidateAccessToken(ctx context.Context, token string) (string, error)
}

type UserInfo struct {
	GUID      string `json:"guid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
