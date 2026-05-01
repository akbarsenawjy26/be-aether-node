package installation_point

import "time"

type InstallationPoint struct {
	GUID         string     `json:"guid"`
	Name         string     `json:"name"`
	DeviceGUID   string     `json:"device_guid"`
	LocationGUID string     `json:"location_guid"`
	Notes        string     `json:"notes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// InstallationPointWithRelations includes device and location details
type InstallationPointWithRelations struct {
	InstallationPoint
	DeviceSerialNumber string `json:"device_serial_number,omitempty"`
	DeviceAlias       string `json:"device_alias,omitempty"`
	LocationName      string `json:"location_name,omitempty"`
}
