package apikey

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyInvalid = errors.New("invalid API key")
)

type APIKey struct {
	GUID       string     `json:"guid"`
	KeyHash    string     `json:"-"`
	Notes      string     `json:"notes"`
	ExpireDate time.Time  `json:"expire_date"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type APIKeyWithKey struct {
	APIKey
	Key string `json:"key,omitempty"`
}

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

type CreateAPIKeyRequest struct {
	ExpireDate string
	Notes      string
	IsActive   bool
}

type UpdateAPIKeyRequest struct {
	ExpireDate *string
	Notes      *string
	IsActive   *bool
}

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByGUID(ctx context.Context, guid string) (*APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, guid string) error
	ValidateKey(ctx context.Context, key string) (*APIKey, error)
}

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithKey, error)
	GetAPIKey(ctx context.Context, guid string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateAPIKey(ctx context.Context, guid string, req *UpdateAPIKeyRequest) (*APIKey, error)
	DeleteAPIKey(ctx context.Context, guid string) error
	ValidateAPIKey(ctx context.Context, key string) (*APIKey, error)
}
