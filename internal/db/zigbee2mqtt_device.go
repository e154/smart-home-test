// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2023, Filippov Alex
//
// This library is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Library General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library.  If not, see
// <https://www.gnu.org/licenses/>.

package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Zigbee2mqttDevices ...
type Zigbee2mqttDevices struct {
	*Common
}

// Zigbee2mqttDevice ...
type Zigbee2mqttDevice struct {
	Id            string `gorm:"primary_key"`
	Zigbee2mqtt   *Zigbee2mqtt
	Zigbee2mqttId int64
	Name          string
	Type          string
	Model         string
	Description   string
	Manufacturer  string
	Status        string
	Functions     pq.StringArray  `gorm:"type:varchar(100)[]"`
	Payload       json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt     time.Time       `gorm:"<-:create"`
	UpdatedAt     time.Time
}

// TableName ...
func (m *Zigbee2mqttDevice) TableName() string {
	return "zigbee2mqtt_devices"
}

// Add ...
func (z Zigbee2mqttDevices) Add(ctx context.Context, v *Zigbee2mqttDevice) (err error) {
	if err = z.DB(ctx).Create(&v).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceAdd)
	}
	return
}

// GetById ...
func (z Zigbee2mqttDevices) GetById(ctx context.Context, id string) (v *Zigbee2mqttDevice, err error) {
	v = &Zigbee2mqttDevice{Id: id}
	if err = z.DB(ctx).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("id \"%s\"", id), apperr.ErrZigbeeDeviceNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceGet)
	}
	return
}

// Update ...
func (z Zigbee2mqttDevices) Update(ctx context.Context, m *Zigbee2mqttDevice) (err error) {
	err = z.DB(ctx).Model(&Zigbee2mqttDevice{Id: m.Id}).Updates(map[string]interface{}{
		"Name":         m.Name,
		"Type":         m.Type,
		"Model":        m.Model,
		"Description":  m.Description,
		"Manufacturer": m.Manufacturer,
		"Functions":    m.Functions,
		"Status":       m.Status,
		"Payload":      m.Payload,
	}).Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceUpdate)
	}
	return
}

// Delete ...
func (z Zigbee2mqttDevices) Delete(ctx context.Context, id string) (err error) {
	if err = z.DB(ctx).Delete(&Zigbee2mqttDevice{Id: id}).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceDelete)
	}
	return
}

// List ...
func (z *Zigbee2mqttDevices) List(ctx context.Context, limit, offset int) (list []*Zigbee2mqttDevice, total int64, err error) {

	if err = z.DB(ctx).Model(Zigbee2mqttDevice{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceList)
		return
	}

	list = make([]*Zigbee2mqttDevice, 0)
	err = z.DB(ctx).
		Limit(limit).
		Offset(offset).
		Find(&list).
		Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceList)
	}
	return
}

// ListByBridgeId ...
func (z *Zigbee2mqttDevices) ListByBridgeId(ctx context.Context, bridgeId int64, limit, offset int, orderBy, sort string) (list []*Zigbee2mqttDevice, total int64, err error) {

	if err = z.DB(ctx).Model(Zigbee2mqttDevice{}).Where("zigbee2mqtt_id = ?", bridgeId).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceList)
		return
	}

	list = make([]*Zigbee2mqttDevice, 0)
	q := z.DB(ctx).
		Where("zigbee2mqtt_id = ?", bridgeId).
		Limit(limit).
		Offset(offset)

	if sort != "" && orderBy != "" {
		q = q.Order(fmt.Sprintf("%s %s", sort, orderBy))
	}

	err = q.Find(&list).Error

	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceList)
	}
	return
}

// Search ...
func (z *Zigbee2mqttDevices) Search(ctx context.Context, query string, limit, offset int) (list []*Zigbee2mqttDevice, total int64, err error) {

	q := z.DB(ctx).Model(&Zigbee2mqttDevice{}).
		Where("name ILIKE ?", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceSearch)
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Zigbee2mqttDevice, 0)
	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrZigbeeDeviceSearch)
	}

	return
}
