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
	"strings"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
)

// UserDevices ...
type UserDevices struct {
	*Common
}

// UserDevice ...
type UserDevice struct {
	Id               int64 `gorm:"primary_key"`
	UserId           int64
	PushRegistration json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt        time.Time       `gorm:"<-:create"`
}

// TableName ...
func (d *UserDevice) TableName() string {
	return "user_devices"
}

// Add ...
func (d *UserDevices) Add(ctx context.Context, device *UserDevice) (id int64, err error) {
	if err = d.DB(ctx).Create(&device).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "push_registration_at_user_devices_unq") {
					err = fmt.Errorf("%s: %w", fmt.Sprintf("device \"%s\" not unique", device.PushRegistration), apperr.ErrUserDeviceAdd)
					return
				}
			default:
				fmt.Printf("unknown code \"%s\"\n", pgErr.Code)
			}
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserDeviceAdd)
		return
	}
	id = device.Id
	return
}

// GetByUserId ...
func (d *UserDevices) GetByUserId(ctx context.Context, id int64) (devices []*UserDevice, err error) {
	devices = make([]*UserDevice, 0)
	err = d.DB(ctx).Model(&UserDevice{}).
		Where("user_id = ?", id).
		Find(&devices).
		Error

	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserDeviceGet)
	}
	return
}

// Delete ...
func (d *UserDevices) Delete(ctx context.Context, id int64) (err error) {
	if err = d.DB(ctx).Delete(&UserDevice{}, "id = ?", id).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserDeviceDelete)
	}
	return
}

// List ...
func (d *UserDevices) List(ctx context.Context, limit, offset int, orderBy, sort string) (list []*UserDevice, total int64, err error) {

	list = make([]*UserDevice, 0)
	q := d.DB(ctx).Model(UserDevice{})
	if err = q.Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserDeviceList)
		return
	}
	err = q.
		Limit(limit).
		Offset(offset).
		//Order(fmt.Sprintf("%s %s", sort, orderBy)).
		Find(&list).
		Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserDeviceList)
	}
	return
}
