package models

type DeviceActionScript struct {
	Id int64 `json:"id"`
}
type DeviceActionDevice struct {
	Id int64 `json:"id"`
}

// swagger:model
type NewDeviceAction struct {
	Name        string              `json:"name" valid:"MaxSize(254);Required"`
	Description string              `json:"description"`
	Device      *DeviceActionDevice `json:"device"`
	Script      *DeviceActionScript `json:"script"`
}

// swagger:model
type UpdateDeviceAction struct {
	Id          int64               `json:"id"`
	Name        string              `json:"name" valid:"MaxSize(254);Required"`
	Description string              `json:"description"`
	Device      *DeviceActionDevice `json:"device"`
	Script      *DeviceActionScript `json:"script"`
}

// swagger:model
type DeviceAction struct {
	Id          int64               `json:"id"`
	Name        string              `json:"name" valid:"MaxSize(254);Required"`
	Description string              `json:"description"`
	Device      *DeviceActionDevice `json:"device"`
	DeviceId    int64               `json:"device_id"`
	Script      *DeviceActionScript `json:"script"`
}