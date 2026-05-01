package apikey

import "context"

type CreateAPIKeyRequest struct {
	Notes      string
	ExpireDate string
	IsActive   bool
}

type UpdateAPIKeyRequest struct {
	Notes      *string
	ExpireDate *string
	IsActive   *bool
}

type APIKeyService interface {
	// CreateAPIKey creates a new API key and returns the plain key (only available on creation)
	CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithKey, error)

	// GetAPIKey retrieves an API key by GUID (without the plain key)
	GetAPIKey(ctx context.Context, guid string) (*APIKey, error)

	// ListAPIKeys retrieves API keys with pagination
	ListAPIKeys(ctx context.Context, params *ListParams) (*ListResult, error)

	// UpdateAPIKey updates an existing API key
	UpdateAPIKey(ctx context.Context, guid string, req *UpdateAPIKeyRequest) (*APIKey, error)

	// DeleteAPIKey soft-deletes an API key
	DeleteAPIKey(ctx context.Context, guid string) error

	// ValidateAPIKey checks if an API key is valid
	ValidateAPIKey(ctx context.Context, key string) (*APIKey, error)
}
