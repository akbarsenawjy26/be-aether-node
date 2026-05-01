package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
)

type RefreshToken struct {
	GUID      string    `json:"guid"`
	UserGUID  string    `json:"user_guid"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

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

type UserInfo struct {
	GUID      string `json:"guid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteByUserGUID(ctx context.Context, userGUID string) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context) error
	UpdateLastUsed(ctx context.Context, guid string, usedAt string) error
}

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResult, error)
	Logout(ctx context.Context, userGUID string) error
	Register(ctx context.Context, req *RegisterRequest) (*UserInfo, error)
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResult, error)
	ValidateAccessToken(ctx context.Context, tokenString string) (string, error)
}
