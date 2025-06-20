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
	"gorm.io/gorm"
)

// Triggers ...
type Triggers struct {
	*Common
}

// Trigger ...
type Trigger struct {
	Id          int64 `gorm:"primary_key"`
	Name        string
	Entities    []*Entity `gorm:"many2many:trigger_entities;"`
	Script      *Script
	ScriptId    *int64
	PluginName  string
	Payload     string
	Enabled     bool
	AreaId      *int64
	Area        *Area
	Description string
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// TableName ...
func (*Trigger) TableName() string {
	return "triggers"
}

// Add ...
func (t Triggers) Add(ctx context.Context, trigger *Trigger) (id int64, err error) {
	if err = t.DB(ctx).
		Omit("Entities.*").
		Create(&trigger).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerAdd)
		return
	}
	id = trigger.Id
	return
}

// GetById ...
func (t Triggers) GetById(ctx context.Context, id int64) (trigger *Trigger, err error) {
	trigger = &Trigger{}
	err = t.DB(ctx).Model(trigger).
		Where("id = ?", id).
		Preload("Entities").
		Preload("Script").
		Preload("Area").
		First(&trigger).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("id \"%d\"", id), apperr.ErrTriggerNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerGet)
	}

	return
}

// Update ...
func (t Triggers) Update(ctx context.Context, trigger *Trigger) (err error) {
	err = t.DB(ctx).
		Omit("Entities.*").
		Save(trigger).Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerUpdate)
	}
	return
}

// Delete ...
func (t Triggers) Delete(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Delete(&Trigger{}, "id = ?", id).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerDelete)
	}
	return
}

// List ...
func (t Triggers) List(ctx context.Context, limit, offset int, orderBy, sort string, onlyEnabled bool) (list []*Trigger, total int64, err error) {

	if err = t.DB(ctx).Model(Trigger{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerList)
		return
	}

	list = make([]*Trigger, 0)
	q := t.DB(ctx).Model(&Trigger{})

	if onlyEnabled {
		q = q.Where("enabled = ?", true)
	}

	q = q.
		Preload("Entities").
		Preload("Script").
		Preload("Area").
		Limit(limit).
		Offset(offset)

	if sort != "" && orderBy != "" {
		q = q.
			Order(fmt.Sprintf("%s %s", sort, orderBy))
	}

	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerList)
	}
	return
}

// ListPlain ...
func (t Triggers) ListPlain(ctx context.Context, limit, offset int, orderBy, sort string, onlyEnabled bool, ids *[]uint64) (list []*Trigger, total int64, err error) {

	if err = t.DB(ctx).Model(Trigger{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerList)
		return
	}

	list = make([]*Trigger, 0)
	q := t.DB(ctx).Model(&Trigger{})

	if onlyEnabled {
		q = q.Where("enabled = ?", true)
	}

	q = q.
		Preload("Area").
		Limit(limit).
		Offset(offset)

	if sort != "" && orderBy != "" {
		q = q.
			Order(fmt.Sprintf("%s %s", sort, orderBy))
	}
	if ids != nil {
		q = q.Where("id IN (?)", *ids)
	}
	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerList)
	}
	return
}

// Search ...
func (t Triggers) Search(ctx context.Context, query string, limit, offset int) (list []*Trigger, total int64, err error) {

	q := t.DB(ctx).Model(&Trigger{}).
		Where("name ILIKE ?", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerSearch)
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Trigger, 0)
	err = q.Find(&list).Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerSearch)
	}
	return
}

// Enable ...
func (t Triggers) Enable(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Model(&Trigger{Id: id}).Updates(map[string]interface{}{"enabled": true}).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerUpdate)
		return
	}
	return
}

// Disable ...
func (t Triggers) Disable(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Model(&Trigger{Id: id}).Updates(map[string]interface{}{"enabled": false}).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerUpdate)
		return
	}
	return
}

// DeleteEntity ...
func (t Triggers) DeleteEntity(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Model(&Trigger{Id: id}).Association("Entities").Clear(); err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrTriggerDeleteEntity)
	}
	return
}
