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
	"fmt"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Plugins ...
type Plugins struct {
	*Common
}

// Plugin ...
type Plugin struct {
	Name     string `gorm:"primary_key"`
	Version  string
	Enabled  bool
	System   bool
	Actor    bool
	Triggers bool
	External bool
	Settings json.RawMessage `gorm:"type:jsonb;not null"`
}

// TableName ...
func (d Plugin) TableName() string {
	return "plugins"
}

// Add ...
func (n Plugins) Add(ctx context.Context, plugin *Plugin) (err error) {
	if err = n.DB(ctx).Create(&plugin).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginAdd, err.Error())
		return
	}
	return
}

// CreateOrUpdate ...
func (n Plugins) CreateOrUpdate(ctx context.Context, v *Plugin) (err error) {
	err = n.DB(ctx).Model(&Plugin{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"version":  v.Version,
			"enabled":  v.Enabled,
			"system":   v.System,
			"settings": v.Settings,
			"actor":    v.Actor,
			"triggers": v.Triggers,
			"external": v.External,
		}),
	}).Create(&v).Error
	if err != nil {
		err = errors.Wrap(apperr.ErrPluginUpdate, err.Error())
	}
	return
}

// Update ...
func (n Plugins) Update(ctx context.Context, m *Plugin) (err error) {
	if err = n.DB(ctx).Model(&Plugin{}).Where("name = ?", m.Name).Updates(map[string]interface{}{
		"enabled":  m.Enabled,
		"system":   m.System,
		"actor":    m.Actor,
		"triggers": m.Triggers,
		"settings": m.Settings,
		"version":  m.Version,
		"external": m.External,
	}).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginUpdate, err.Error())
	}
	return
}

// Delete ...
func (n Plugins) Delete(ctx context.Context, name string) (err error) {
	if err = n.DB(ctx).Delete(&Plugin{Name: name}).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginDelete, err.Error())
	}
	return
}

// List ...
func (n Plugins) List(ctx context.Context, limit, offset int, orderBy, sort string, enabled, triggers *bool) (list []*Plugin, total int64, err error) {

	list = make([]*Plugin, 0)
	q := n.DB(ctx).Model(&Plugin{})

	if enabled != nil {
		q = q.
			Where("enabled = ?", *enabled)
	}

	if triggers != nil {
		q = q.Where("triggers = ?", *triggers)
	}

	if err = q.Count(&total).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginList, err.Error())
		return
	}

	if sort != "" && orderBy != "" {
		q = q.
			Order(fmt.Sprintf("%s %s", sort, orderBy))
	}

	q = q.
		Limit(limit).
		Offset(offset)

	if err = q.Find(&list).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginList, err.Error())
	}

	return
}

// Search ...
func (n Plugins) Search(ctx context.Context, query string, limit, offset int) (list []*Plugin, total int64, err error) {

	q := n.DB(ctx).Model(&Plugin{}).
		Where("name LIKE ? and actor=true and enabled=true", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginSearch, err.Error())
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Plugin, 0)
	if err = q.Find(&list).Error; err != nil {
		err = errors.Wrap(apperr.ErrPluginSearch, err.Error())
	}

	return
}

// GetByName ...
func (n Plugins) GetByName(ctx context.Context, name string) (plugin *Plugin, err error) {

	plugin = &Plugin{}
	err = n.DB(ctx).Model(plugin).
		Where("name = ?", name).
		First(&plugin).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(apperr.ErrPluginNotFound, fmt.Sprintf("name \"%s\"", name))
			return
		}
		err = errors.Wrap(apperr.ErrPluginGet, err.Error())
	}
	return
}
