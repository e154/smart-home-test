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
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Conditions ...
type Conditions struct {
	*Common
}

// Condition ...
type Condition struct {
	Id          int64 `gorm:"primary_key"`
	Name        string
	Script      *Script
	ScriptId    *int64
	AreaId      *int64
	Area        *Area
	Description string
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// TableName ...
func (d *Condition) TableName() string {
	return "conditions"
}

// Add ...
func (t Conditions) Add(ctx context.Context, condition *Condition) (id int64, err error) {
	if err = t.DB(ctx).Create(&condition).Error; err != nil {
		err = errors.Wrap(apperr.ErrConditionAdd, err.Error())
		return
	}
	id = condition.Id
	return
}

// GetById ...
func (t Conditions) GetById(ctx context.Context, id int64) (condition *Condition, err error) {
	condition = &Condition{Id: id}
	err = t.DB(ctx).
		WithContext(ctx).
		Model(condition).
		Preload("Script").
		Preload("Area").
		First(&condition).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(apperr.ErrConditionNotFound, fmt.Sprintf("id \"%d\"", id))
			return
		}
		err = errors.Wrap(apperr.ErrConditionGet, err.Error())
	}
	return
}

// Update ...
func (t Conditions) Update(ctx context.Context, m *Condition) (err error) {

	q := map[string]interface{}{
		"name":        m.Name,
		"description": m.Description,
		"script_id":   m.ScriptId,
		"area_id":     m.AreaId,
	}

	if err = t.DB(ctx).Model(&Condition{}).Where("id = ?", m.Id).Updates(q).Error; err != nil {
		err = errors.Wrap(apperr.ErrConditionUpdate, err.Error())
	}
	return
}

// Delete ...
func (t Conditions) Delete(ctx context.Context, id int64) (err error) {
	if err = t.DB(ctx).Delete(&Condition{}, "id = ?", id).Error; err != nil {
		err = errors.Wrap(apperr.ErrConditionDelete, err.Error())
	}
	return
}

// List ...
func (t *Conditions) List(ctx context.Context, limit, offset int, orderBy, sort string, ids *[]uint64) (list []*Condition, total int64, err error) {

	if err = t.DB(ctx).Model(Condition{}).Count(&total).Error; err != nil {
		err = errors.Wrap(apperr.ErrConditionList, err.Error())
		return
	}

	list = make([]*Condition, 0)
	q := t.DB(ctx).Model(&Condition{}).
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
		err = errors.Wrap(apperr.ErrConditionList, err.Error())
	}
	return
}

// Search ...q
func (t *Conditions) Search(ctx context.Context, query string, limit, offset int) (list []*Condition, total int64, err error) {

	q := t.DB(ctx).Model(&Condition{}).
		Where("name LIKE ?", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = errors.Wrap(apperr.ErrConditionSearch, err.Error())
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Condition, 0)
	err = q.Find(&list).Error
	if err != nil {
		err = errors.Wrap(apperr.ErrConditionSearch, err.Error())
	}
	return
}
