package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type apiKeyService struct {
	repo     APIKeyRepository
	keyPrefix string
}

func NewAPIKeyService(repo APIKeyRepository, keyPrefix string) APIKeyService {
	if keyPrefix == "" {
		keyPrefix = "aeth_live_pk_"
	}
	return &apiKeyService{repo: repo, keyPrefix: keyPrefix}
}

func (s *apiKeyService) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithKey, error) {
	// Generate a random key
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	plainKey := s.keyPrefix + hex.EncodeToString(randomBytes)

	// Hash the key for storage
	keyHash := hashKey(plainKey)

	// Parse expire date
	expireDate, err := time.Parse(time.RFC3339, req.ExpireDate)
	if err != nil {
		return nil, err
	}

	apiKey := &APIKey{
		KeyHash:    keyHash,
		Notes:      req.Notes,
		ExpireDate: expireDate,
		IsActive:   req.IsActive,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	// Return with the plain key (only available on creation)
	return &APIKeyWithKey{
		APIKey: *apiKey,
		Key:    plainKey,
	}, nil
}

func (s *apiKeyService) GetAPIKey(ctx context.Context, guid string) (*APIKey, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *apiKeyService) ListAPIKeys(ctx context.Context, params *ListParams) (*ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *apiKeyService) UpdateAPIKey(ctx context.Context, guid string, req *UpdateAPIKeyRequest) (*APIKey, error) {
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

func (s *apiKeyService) ValidateAPIKey(ctx context.Context, key string) (*APIKey, error) {
	return s.repo.ValidateKey(ctx, key)
}

func hashKey(key string) string {
	// Simple hash for storage - in production use crypto/sha256
	// This is already done in repository with proper hashing
	return key
}
