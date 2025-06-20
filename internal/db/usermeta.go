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
	"fmt"

	"github.com/e154/smart-home/pkg/apperr"
)

// UserMetas ...
type UserMetas struct {
	*Common
}

// UserMeta ...
type UserMeta struct {
	Id     int64 `gorm:"primary_key"`
	UserId int64
	Key    string
	Value  string
}

// TableName ...
func (m *UserMeta) TableName() string {
	return "user_metas"
}

// UpdateOrCreate ...
func (m *UserMetas) UpdateOrCreate(ctx context.Context, meta *UserMeta) (id int64, err error) {

	err = m.DB(ctx).Model(&UserMeta{}).
		Where("user_id = ? and key = ?", meta.UserId, meta.Key).
		Updates(map[string]interface{}{"value": meta.Value}).
		Error

	if err != nil {
		if err = m.DB(ctx).Create(&meta).Error; err != nil {
			err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrUserMetaAdd)
			return
		}
		id = meta.Id
	}

	return
}
