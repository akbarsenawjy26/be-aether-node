package auth

import (
	"context"
	"time"
)

type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// DeleteByUserGUID deletes all refresh tokens for a user
	DeleteByUserGUID(ctx context.Context, userGUID string) error

	// DeleteByTokenHash deletes a specific refresh token
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// DeleteExpired deletes all expired refresh tokens
	DeleteExpired(ctx context.Context) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, guid string, usedAt time.Time) error
}
