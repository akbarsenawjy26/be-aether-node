package user

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type User struct {
	GUID         string     `json:"guid"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	RoleGUID     *string    `json:"role_guid,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

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

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByGUID(ctx context.Context, guid string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, guid string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateLastLogin(ctx context.Context, guid string, loginAt time.Time) error
}

type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, guid string) (*User, error)
	ListUsers(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateUser(ctx context.Context, guid string, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, guid string) error
}
