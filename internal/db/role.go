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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"gorm.io/gorm"
)

// Roles ...
type Roles struct {
	*Common
}

// Role ...
type Role struct {
	Name        string `gorm:"primary_key"`
	Description string
	Role        *Role
	RoleName    sql.NullString `gorm:"column:parent"`
	Children    []*Role
	Permissions []*Permission
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// TableName ...
func (m *Role) TableName() string {
	return "roles"
}

// Add ...
func (n Roles) Add(ctx context.Context, role *Role) (err error) {
	if err = n.DB(ctx).Create(&role).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleAdd)
		return
	}
	return
}

// GetByName ...
func (n Roles) GetByName(ctx context.Context, name string) (role *Role, err error) {
	role = &Role{Name: name}
	err = n.DB(ctx).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("name \"%s\"", name), apperr.ErrRoleNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleGet)
		return
	}

	err = n.RelData(ctx, role)

	return
}

// Update ...
func (n Roles) Update(ctx context.Context, m *Role) (err error) {
	err = n.DB(ctx).Model(&Role{Name: m.Name}).Updates(map[string]interface{}{
		"description": m.Description,
		"parent":      m.RoleName,
	}).Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleUpdate)
	}

	return
}

// Delete ...
func (n Roles) Delete(ctx context.Context, name string) (err error) {
	if err = n.DB(ctx).Delete(&Role{Name: name}).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleDelete)
	}
	return
}

// List ...
func (n *Roles) List(ctx context.Context, limit, offset int, orderBy, sort string) (list []*Role, total int64, err error) {

	if err = n.DB(ctx).Model(Role{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleList)
		return
	}

	list = make([]*Role, 0)
	err = n.DB(ctx).
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", sort, orderBy)).
		Find(&list).
		Error

	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleList)
		return
	}

	for _, role := range list {
		_ = n.RelData(ctx, role)
	}

	return
}

// Search ...
func (n *Roles) Search(ctx context.Context, query string, limit, offset int) (list []*Role, total int64, err error) {

	q := n.DB(ctx).Model(&Role{}).
		Where("name ILIKE ?", "%"+query+"%")

	if err = q.Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleSearch)
		return
	}

	q = q.
		Limit(limit).
		Offset(offset).
		Order("name ASC")

	list = make([]*Role, 0)
	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleSearch)
	}

	return
}

// RelData ...
func (n *Roles) RelData(ctx context.Context, role *Role) (err error) {

	// get parent
	if role.RoleName.Valid {
		role.Role = &Role{}
		err = n.DB(ctx).Model(role).
			Where("name = ?", role.RoleName.String).
			Find(&role.Role).
			Error
		if err != nil {
			err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleGet)
		}
	}

	// get children
	role.Children = make([]*Role, 0)
	err = n.DB(ctx).Model(role).
		Where("parent = ?", role.Name).
		Find(&role.Children).
		Error
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrRoleGet)
	}

	return
}
