package location

import "time"

type Location struct {
	GUID      string     `json:"guid"`
	Name      string     `json:"name"`
	Notes     string     `json:"notes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
