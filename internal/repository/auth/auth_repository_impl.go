package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"aether-node/internal/db"
	domainAuth "aether-node/internal/domain/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type refreshTokenRepository struct {
	db *db.Queries
}

func NewRefreshTokenRepository(queries *db.Queries) domainAuth.RefreshTokenRepository {
	return &refreshTokenRepository{db: queries}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domainAuth.RefreshToken) error {
	if token.GUID == "" {
		token.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(token.GUID)
	userGUID, _ := uuid.Parse(token.UserGUID)
	now := time.Now()

	params := db.CreateRefreshTokenParams{
		Guid:      db.NewUUID(guid),
		UserGuid:  db.NewUUID(userGUID),
		TokenHash: token.TokenHash,
		ExpiresAt: db.NewTimestamptz(token.ExpiresAt),
		CreatedAt: db.NewTimestamptz(now),
	}

	return r.db.CreateRefreshToken(ctx, params)
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domainAuth.RefreshToken, error) {
	dbToken, err := r.db.GetRefreshTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	rt := db.RefreshTokenFromDB(&dbToken)

	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	return rt, nil
}

func (r *refreshTokenRepository) DeleteByUserGUID(ctx context.Context, userGUID string) error {
	id, _ := uuid.Parse(userGUID)
	return r.db.DeleteRefreshTokensByUserGUID(ctx, db.NewUUID(id))
}

func (r *refreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return r.db.DeleteRefreshTokenByTokenHash(ctx, tokenHash)
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.DeleteExpiredRefreshTokens(ctx, db.NewTimestamptz(time.Now()))
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, guid string, usedAt string) error {
	return nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
