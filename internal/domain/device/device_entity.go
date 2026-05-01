package device

import "time"

type Device struct {
	GUID         string     `json:"guid"`
	Type         string     `json:"type"`
	SerialNumber string     `json:"serial_number"`
	Alias        string     `json:"alias"`
	Notes        string     `json:"notes"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
