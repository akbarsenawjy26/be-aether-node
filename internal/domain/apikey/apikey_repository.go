package apikey

import (
	"context"
	"time"
)

type ListParams struct {
	Limit  int
	Page   int
	Order  string
	Sort   string
	Search string
}

type ListResult struct {
	APIKeys   []*APIKey
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type APIKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, apiKey *APIKey) error

	// GetByGUID retrieves an API key by GUID
	GetByGUID(ctx context.Context, guid string) (*APIKey, error)

	// GetByKeyHash retrieves an API key by its hash
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)

	// List retrieves API keys with pagination
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// Update updates an existing API key
	Update(ctx context.Context, apiKey *APIKey) error

	// Delete soft-deletes an API key
	Delete(ctx context.Context, guid string) error

	// ValidateKey checks if an API key is valid (exists, active, not expired)
	ValidateKey(ctx context.Context, key string) (*APIKey, error)
}
