package apikey

import "time"

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

// APIKeyWithKey includes the plain API key (only on creation)
type APIKeyWithKey struct {
	APIKey
	Key string `json:"key,omitempty"`
}
