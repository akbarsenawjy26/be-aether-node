package user

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
	Users     []*User
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByGUID retrieves a user by GUID
	GetByGUID(ctx context.Context, guid string) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// List retrieves users with pagination
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete soft-deletes a user (sets deleted_at)
	Delete(ctx context.Context, guid string) error

	// ExistsByEmail checks if email already exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, guid string, loginAt time.Time) error
}
