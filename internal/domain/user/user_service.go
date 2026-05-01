package user

import "context"

type CreateUserRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleGUID  *string
}

type UpdateUserRequest struct {
	Email     *string
	Password  *string
	FirstName *string
	LastName  *string
	RoleGUID  *string
	IsActive  *bool
}

type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)

	// GetUser retrieves a user by GUID
	GetUser(ctx context.Context, guid string) (*User, error)

	// ListUsers retrieves users with pagination
	ListUsers(ctx context.Context, params *ListParams) (*ListResult, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, guid string, req *UpdateUserRequest) (*User, error)

	// DeleteUser soft-deletes a user
	DeleteUser(ctx context.Context, guid string) error
}
