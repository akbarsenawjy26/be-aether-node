package apikey

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"aether-node/internal/db"
	domainAPIKey "aether-node/internal/domain/apikey"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyInvalid = errors.New("invalid API key")
)

type apiKeyRepository struct {
	db *db.Queries
}

func NewAPIKeyRepository(pool *pgxpool.Pool) domainAPIKey.APIKeyRepository {
	return &apiKeyRepository{db: db.New(pool)}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *domainAPIKey.APIKey) error {
	if apiKey.GUID == "" {
		apiKey.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(apiKey.GUID)
	now := time.Now()

	params := db.CreateAPIKeyParams{
		Guid:       db.NewUUID(guid),
		KeyHash:    apiKey.KeyHash,
		ExpireDate: db.NewTimestamptz(apiKey.ExpireDate),
		IsActive:   apiKey.IsActive,
		CreatedAt:  db.NewTimestamptz(now),
		UpdatedAt:  db.NewTimestamptz(now),
	}

	if apiKey.Notes != "" {
		params.Notes = db.NewText(apiKey.Notes)
	}

	return r.db.CreateAPIKey(ctx, params)
}

func (r *apiKeyRepository) GetByGUID(ctx context.Context, guid string) (*domainAPIKey.APIKey, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	dbKey, err := r.db.GetAPIKeyByGUID(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return db.APIKeyFromDB(&dbKey), nil
}

func (r *apiKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domainAPIKey.APIKey, error) {
	dbKey, err := r.db.GetAPIKeyByKeyHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return db.APIKeyFromDB(&dbKey), nil
}

func (r *apiKeyRepository) List(ctx context.Context, params domainAPIKey.ListParams) (*domainAPIKey.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	search := "%" + params.Search + "%"
	offset := (params.Page - 1) * params.Limit

	dbKeys, err := r.db.ListAPIKeys(ctx, db.ListAPIKeysParams{
		Notes:  db.NewText(search),
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	apiKeys := make([]*domainAPIKey.APIKey, 0, len(dbKeys))
	for i := range dbKeys {
		apiKeys = append(apiKeys, db.APIKeyFromDB(&dbKeys[i]))
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainAPIKey.ListResult{
		APIKeys:   apiKeys,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *domainAPIKey.APIKey) error {
	guid, _ := uuid.Parse(apiKey.GUID)
	now := time.Now()

	params := db.UpdateAPIKeyParams{
		Guid:       db.NewUUID(guid),
		ExpireDate: db.NewTimestamptz(apiKey.ExpireDate),
		IsActive:   apiKey.IsActive,
		UpdatedAt:  db.NewTimestamptz(now),
	}

	if apiKey.Notes != "" {
		params.Notes = db.NewText(apiKey.Notes)
	}

	err := r.db.UpdateAPIKey(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAPIKeyNotFound
		}
		return err
	}
	return nil
}

func (r *apiKeyRepository) Delete(ctx context.Context, guid string) error {
	id, err := uuid.Parse(guid)
	if err != nil {
		return ErrAPIKeyNotFound
	}

	err = r.db.DeleteAPIKey(ctx, db.DeleteAPIKeyParams{
		Guid:      db.NewUUID(id),
		DeletedAt: db.NewTimestamptz(time.Now()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAPIKeyNotFound
		}
		return err
	}
	return nil
}

func (r *apiKeyRepository) ValidateKey(ctx context.Context, key string) (*domainAPIKey.APIKey, error) {
	keyHash := sha256.Sum256([]byte(key))
	keyHashStr := hex.EncodeToString(keyHash[:])

	apiKey, err := r.GetByKeyHash(ctx, keyHashStr)
	if err != nil {
		if errors.Is(err, ErrAPIKeyNotFound) {
			return nil, ErrAPIKeyInvalid
		}
		return nil, err
	}

	if !apiKey.IsActive {
		return nil, ErrAPIKeyInvalid
	}

	if time.Now().After(apiKey.ExpireDate) {
		return nil, ErrAPIKeyInvalid
	}

	return apiKey, nil
}
