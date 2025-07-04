// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2024, Filippov Alex
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

package adaptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/e154/smart-home/internal/db"
	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/apperr"
	pkgCommon "github.com/e154/smart-home/pkg/common"
	m "github.com/e154/smart-home/pkg/models"

	"gorm.io/gorm"
)

var _ adaptors.EntityActionRepo = (*EntityAction)(nil)

// EntityAction ...
type EntityAction struct {
	table *db.EntityActions
	db    *gorm.DB
}

// GetEntityActionAdaptor ...
func GetEntityActionAdaptor(d *gorm.DB) *EntityAction {
	return &EntityAction{
		table: &db.EntityActions{&db.Common{Db: d}},
		db:    d,
	}
}

// Add ...
func (n *EntityAction) Add(ctx context.Context, ver *m.EntityAction) (id int64, err error) {

	dbVer := n.toDb(ver)
	if id, err = n.table.Add(ctx, dbVer); err != nil {
		return
	}

	return
}

// DeleteByEntityId ...
func (n *EntityAction) DeleteByEntityId(ctx context.Context, id pkgCommon.EntityId) (err error) {
	err = n.table.DeleteByEntityId(ctx, id)
	return
}

// AddMultiple ...
func (n *EntityAction) AddMultiple(ctx context.Context, items []*m.EntityAction) (err error) {

	if len(items) == 0 {
		return
	}

	insertRecords := make([]*db.EntityAction, 0, len(items))

	for _, ver := range items {
		//if ver.ImageId == 0 {
		//	continue
		//}
		insertRecords = append(insertRecords, n.toDb(ver))
	}

	if err = n.table.AddMultiple(ctx, insertRecords); err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrEntityActionAdd)
	}

	return
}

func (n *EntityAction) fromDb(dbVer *db.EntityAction) (ver *m.EntityAction) {
	ver = &m.EntityAction{
		Id:          dbVer.Id,
		Name:        dbVer.Name,
		Description: dbVer.Description,
		Icon:        dbVer.Icon,
		EntityId:    dbVer.EntityId,
		ImageId:     dbVer.ImageId,
		ScriptId:    dbVer.ScriptId,
		Type:        dbVer.Type,
		CreatedAt:   dbVer.CreatedAt,
		UpdatedAt:   dbVer.UpdatedAt,
	}

	// image
	if dbVer.Image != nil {
		imageAdaptor := GetImageAdaptor(n.db)
		ver.Image = imageAdaptor.fromDb(dbVer.Image)
	}

	// script
	if dbVer.Script != nil {
		scriptAdaptor := GetScriptAdaptor(n.db)
		ver.Script, _ = scriptAdaptor.fromDb(dbVer.Script)
	}

	return
}

func (n *EntityAction) toDb(ver *m.EntityAction) (dbVer *db.EntityAction) {
	dbVer = &db.EntityAction{
		Id:          ver.Id,
		Name:        strings.TrimSpace(ver.Name),
		Description: ver.Description,
		Icon:        ver.Icon,
		EntityId:    ver.EntityId,
		ImageId:     ver.ImageId,
		ScriptId:    ver.ScriptId,
		Type:        ver.Type,
	}
	if ver.Image != nil && ver.Image.Id != 0 {
		dbVer.ImageId = pkgCommon.Int64(ver.Image.Id)
	}
	if ver.Script != nil && ver.Script.Id != 0 {
		dbVer.ScriptId = pkgCommon.Int64(ver.Script.Id)
	}
	return
}
