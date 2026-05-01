package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	domainAPIKey "aether-node/internal/domain/apikey"
)

type apiKeyService struct {
	repo      domainAPIKey.APIKeyRepository
	keyPrefix string
}

func NewAPIKeyService(repo domainAPIKey.APIKeyRepository, keyPrefix string) domainAPIKey.APIKeyService {
	if keyPrefix == "" {
		keyPrefix = "aeth_live_pk_"
	}
	return &apiKeyService{repo: repo, keyPrefix: keyPrefix}
}

func (s *apiKeyService) CreateAPIKey(ctx context.Context, req *domainAPIKey.CreateAPIKeyRequest) (*domainAPIKey.APIKeyWithKey, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	plainKey := s.keyPrefix + hex.EncodeToString(randomBytes)

	keyHash := hashKey(plainKey)

	expireDate, err := time.Parse(time.RFC3339, req.ExpireDate)
	if err != nil {
		return nil, err
	}

	apiKey := &domainAPIKey.APIKey{
		KeyHash:    keyHash,
		Notes:      req.Notes,
		ExpireDate: expireDate,
		IsActive:   req.IsActive,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return &domainAPIKey.APIKeyWithKey{
		APIKey: *apiKey,
		Key:    plainKey,
	}, nil
}

func (s *apiKeyService) GetAPIKey(ctx context.Context, guid string) (*domainAPIKey.APIKey, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *apiKeyService) ListAPIKeys(ctx context.Context, params *domainAPIKey.ListParams) (*domainAPIKey.ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *apiKeyService) UpdateAPIKey(ctx context.Context, guid string, req *domainAPIKey.UpdateAPIKeyRequest) (*domainAPIKey.APIKey, error) {
	apiKey, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.Notes != nil {
		apiKey.Notes = *req.Notes
	}

	if req.ExpireDate != nil {
		expireDate, err := time.Parse(time.RFC3339, *req.ExpireDate)
		if err != nil {
			return nil, err
		}
		apiKey.ExpireDate = expireDate
	}

	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *apiKeyService) DeleteAPIKey(ctx context.Context, guid string) error {
	return s.repo.Delete(ctx, guid)
}

func (s *apiKeyService) ValidateAPIKey(ctx context.Context, key string) (*domainAPIKey.APIKey, error) {
	return s.repo.ValidateKey(ctx, key)
}

func hashKey(key string) string {
	return key
}
