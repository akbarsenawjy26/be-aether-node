package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type refreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	if token.GUID == "" {
		token.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO refresh_tokens (guid, user_guid, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		token.GUID,
		token.UserGUID,
		token.TokenHash,
		token.ExpiresAt,
		now,
	)

	return err
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT guid, user_guid, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.GUID,
		&token.UserGUID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("refresh token not found")
	}
	if err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	return token, nil
}

func (r *refreshTokenRepository) DeleteByUserGUID(ctx context.Context, userGUID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_guid = $1`
	_, err := r.db.Exec(ctx, query, userGUID)
	return err
}

func (r *refreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, guid string, usedAt time.Time) error {
	// No-op for now - can be used to track token usage
	return nil
}

// Helper function to hash a token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
