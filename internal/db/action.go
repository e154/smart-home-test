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
	"errors"
	"fmt"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/e154/smart-home/pkg/common"
	"gorm.io/gorm"
)

// Actions ...
type Actions struct {
	*Common
}

// Action ...
type Action struct {
	Id               int64 `gorm:"primary_key"`
	Name             string
	Script           *Script
	ScriptId         *int64
	Entity           *Entity
	EntityId         *common.EntityId
	EntityActionName *string
	AreaId           *int64
	Area             *Area
	Description      string
	CreatedAt        time.Time `gorm:"<-:create"`
	UpdatedAt        time.Time
}

// TableName ...
func (*Action) TableName() string {
	return "actions"
}

// Add ...
func (t Actions) Add(ctx context.Context, action *Action) (id int64, err error) {
	if err = t.DB(ctx).Create(&action).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionAdd)
		return
	}
	id = action.Id
	return
}

// GetById ...
func (t Actions) GetById(ctx context.Context, id int64) (action *Action, err error) {
	action = &Action{}
	err = t.DB(ctx).Model(action).
		Where("id = ?", id).
		Preload("Entity").
		Preload("Script").
		Preload("Area").
		First(&action).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("id \"%d\"", id), apperr.ErrActionNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionGet)
	}
	return
}

// Update ...
func (t Actions) Update(ctx context.Context, m *Action) (err error) {
	q := map[string]interface{}{
		"name":               m.Name,
		"description":        m.Description,
		"script_id":          m.ScriptId,
		"entity_id":          m.EntityId,
		"area_id":            m.AreaId,
		"entity_action_name": m.EntityActionName,
	}
	if err = t.DB(ctx).Model(&Action{}).Where("id = ?", m.Id).Updates(q).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionUpdate)
	}
	return
}

// Delete ...
func (t Actions) Delete(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Delete(&Action{}, "id = ?", id).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionDelete)
	}
	return
}

// List ...
func (t *Actions) List(ctx context.Context, limit, offset int, orderBy, sort string, ids *[]uint64) (list []*Action, total int64, err error) {

	if err = t.DB(ctx).Model(Action{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionList)
		return
	}

	list = make([]*Action, 0)
	q := t.DB(ctx).Model(&Action{}).
		Limit(limit).
		Offset(offset)

	if sort != "" && orderBy != "" {
		q = q.
			Preload("Area").
			Order(fmt.Sprintf("%s %s", sort, orderBy))
	}
	if ids != nil {
		q = q.Where("id IN (?)", *ids)
	}
	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionList)
	}
	return
}

// Search ...q
func (t *Actions) Search(ctx context.Context, query string, limit, offset int) (list []*Action, total int64, err error) {

	q := t.DB(ctx).Model(&Action{}).
		Where("name ILIKE ?", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionSearch)
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Action, 0)
	err = q.Find(&list).Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrActionSearch)
	}
	return
}
