package adaptors

import (
	"github.com/jinzhu/gorm"
	"github.com/e154/smart-home/db"
	m "github.com/e154/smart-home/models"
)

type DeviceState struct {
	table *db.DeviceStates
	db    *gorm.DB
}

func GetDeviceStateAdaptor(d *gorm.DB) *DeviceState {
	return &DeviceState{
		table: &db.DeviceStates{Db: d},
		db:    d,
	}
}

func (n *DeviceState) Add(device *m.DeviceState) (id int64, err error) {

	dbDeviceState := n.toDb(device)
	if id, err = n.table.Add(dbDeviceState); err != nil {
		return
	}

	return
}

func (n *DeviceState) GetById(deviceId int64) (device *m.DeviceState, err error) {

	var dbDeviceState *db.DeviceState
	if dbDeviceState, err = n.table.GetById(deviceId); err != nil {
		return
	}

	device = n.fromDb(dbDeviceState)

	return
}

func (n *DeviceState) Update(device *m.DeviceState) (err error) {
	dbDeviceState := n.toDb(device)
	err = n.table.Update(dbDeviceState)
	return
}

func (n *DeviceState) Delete(deviceId int64) (err error) {
	err = n.table.Delete(deviceId)
	return
}

func (n *DeviceState) List(limit, offset int64, orderBy, sort string) (list []*m.DeviceState, total int64, err error) {
	var dbList []*db.DeviceState
	if dbList, total, err = n.table.List(limit, offset, orderBy, sort); err != nil {
		return
	}

	list = make([]*m.DeviceState, 0)
	for _, dbDeviceState := range dbList {
		device := n.fromDb(dbDeviceState)
		list = append(list, device)
	}

	return
}

func (n *DeviceState) fromDb(dbDeviceState *db.DeviceState) (device *m.DeviceState) {
	device = &m.DeviceState{
		Id:          dbDeviceState.Id,
		Description: dbDeviceState.Description,
		CreatedAt:   dbDeviceState.CreatedAt,
		UpdatedAt:   dbDeviceState.UpdatedAt,
	}
	return
}

func (n *DeviceState) toDb(device *m.DeviceState) (dbDeviceState *db.DeviceState) {
	dbDeviceState = &db.DeviceState{
		Id:          device.Id,
		Description: device.Description,
		CreatedAt:   device.CreatedAt,
		UpdatedAt:   device.UpdatedAt,
	}
	return
}