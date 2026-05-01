package user

import "time"

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
