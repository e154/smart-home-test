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
	"strings"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	pkgCommon "github.com/e154/smart-home/pkg/common"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"errors"
	"gorm.io/gorm"
)

// EntityStates ...
type EntityStates struct {
	*Common
}

// EntityState ...
type EntityState struct {
	Id          int64 `gorm:"primary_key"`
	Name        string
	Description string
	Icon        *string
	Entity      *Entity
	EntityId    pkgCommon.EntityId
	Image       *Image
	ImageId     *int64
	Style       string
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// TableName ...
func (d *EntityState) TableName() string {
	return "entity_states"
}

// Add ...
func (n EntityStates) Add(ctx context.Context, v *EntityState) (id int64, err error) {
	if err = n.DB(ctx).Create(&v).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "name_at_entity_states_unq") {
					err = fmt.Errorf("%s: %w", fmt.Sprintf("state name \"%s\" not unique", v.Name), apperr.ErrEntityStateAdd)
					return
				}
			default:
				fmt.Printf("unknown code \"%s\"\n", pgErr.Code)
			}
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateAdd)
		return
	}
	id = v.Id
	return
}

// GetById ...
func (n EntityStates) GetById(ctx context.Context, id int64) (v *EntityState, err error) {
	v = &EntityState{Id: id}
	if err = n.DB(ctx).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("id \"%d\"", id), apperr.ErrEntityStateNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateGet)
	}
	return
}

// Update ...
func (n EntityStates) Update(ctx context.Context, m *EntityState) (err error) {
	err = n.DB(ctx).Model(&EntityState{Id: m.Id}).Updates(map[string]interface{}{
		"name":        m.Name,
		"description": m.Description,
		"icon":        m.Icon,
		//"device_state_id": m.DeviceStateId,
		"entity_id": m.EntityId,
		"image_id":  m.ImageId,
		"style":     m.Style,
	}).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "name_at_entity_states_unq") {
					err = fmt.Errorf("%s: %w", fmt.Sprintf("state name \"%s\" not unique", m.Name), apperr.ErrEntityStateUpdate)
					return
				}
			default:
				fmt.Printf("unknown code \"%s\"\n", pgErr.Code)
			}
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateUpdate)
	}
	return
}

// DeleteByEntityId ...
func (n EntityStates) DeleteByEntityId(ctx context.Context, entityId pkgCommon.EntityId) (err error) {
	if err = n.DB(ctx).Delete(&EntityState{}, "entity_id = ?", entityId).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateDelete)
		return
	}
	return
}

// List ...
func (n *EntityStates) List(ctx context.Context, limit, offset int, orderBy, sort string) (list []*EntityState, total int64, err error) {

	if err = n.DB(ctx).Model(EntityState{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateList)
		return
	}

	list = make([]*EntityState, 0)
	err = n.DB(ctx).
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", sort, orderBy)).
		Find(&list).
		Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateList)
		return
	}
	return
}

// AddMultiple ...
func (n *EntityStates) AddMultiple(ctx context.Context, states []*EntityState) (err error) {
	if err = n.DB(ctx).Create(&states).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "name_at_entity_states_unq") {
					err = fmt.Errorf("%s: %w", "multiple insert", apperr.ErrEntityStateUpdate)
					return
				}
			default:
				fmt.Printf("unknown code \"%s\"\n", pgErr.Code)
			}
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityStateAdd)
	}
	return
}
