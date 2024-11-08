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

	"github.com/e154/smart-home/internal/db"
	"github.com/e154/smart-home/internal/system/orm"
	"github.com/e154/smart-home/pkg/adaptors"
	m "github.com/e154/smart-home/pkg/models"

	"gorm.io/gorm"
)

var _ adaptors.TaskRepo = (*Task)(nil)

// Task ...
type Task struct {
	table *db.Tasks
	db    *gorm.DB
	orm   *orm.Orm
}

// GetTaskAdaptor ...
func GetTaskAdaptor(d *gorm.DB, orm *orm.Orm) *Task {
	return &Task{
		table: &db.Tasks{&db.Common{Db: d}},
		db:    d,
		orm:   orm,
	}
}

// Import ...
func (n *Task) Import(ctx context.Context, ver *m.Task) (err error) {
	ver.Id, err = n.table.Add(ctx, n.toDb(ver))
	return
}

// Add ...
func (n *Task) Add(ctx context.Context, ver *m.NewTask) (taskId int64, err error) {

	task := &db.Task{
		Name:        ver.Name,
		Description: ver.Description,
		Enabled:     ver.Enabled,
		Condition:   ver.Condition,
		AreaId:      ver.AreaId,
	}

	//triggers
	for _, id := range ver.TriggerIds {
		task.Triggers = append(task.Triggers, &db.Trigger{Id: id})
	}

	//conditions
	for _, id := range ver.ConditionIds {
		task.Conditions = append(task.Conditions, &db.Condition{Id: id})
	}

	//actions
	for _, id := range ver.ActionIds {
		task.Actions = append(task.Actions, &db.Action{Id: id})
	}

	taskId, err = n.table.Add(ctx, task)

	return
}

func (n *Task) DeleteTrigger(ctx context.Context, taskID int64) error {
	return n.table.DeleteTrigger(ctx, taskID)
}

func (n *Task) DeleteCondition(ctx context.Context, taskID int64) error {
	return n.table.DeleteTrigger(ctx, taskID)
}

func (n *Task) DeleteAction(ctx context.Context, taskID int64) error {
	return n.table.DeleteAction(ctx, taskID)
}

// Update ...
func (n *Task) Update(ctx context.Context, task *m.Task) (err error) {
	err = n.table.Update(ctx, n.toDb(task))
	return
}

// Enable ...
func (n *Task) Enable(ctx context.Context, id int64) (err error) {
	err = n.table.Enable(ctx, id)
	return
}

// Disable ...
func (n *Task) Disable(ctx context.Context, id int64) (err error) {
	err = n.table.Disable(ctx, id)
	return
}

// GetById ...
func (n *Task) GetById(ctx context.Context, id int64) (task *m.Task, err error) {

	var dbVer *db.Task
	if dbVer, err = n.table.GetById(ctx, id); err != nil {
		return
	}

	task = n.fromDb(dbVer)

	return
}

// Delete ...
func (n *Task) Delete(ctx context.Context, id int64) (err error) {
	err = n.table.Delete(ctx, id)
	return
}

// List ...
func (n *Task) List(ctx context.Context, limit, offset int64, orderBy, sort string, onlyEnabled bool) (list []*m.Task, total int64, err error) {

	var dbList []*db.Task
	if dbList, total, err = n.table.List(ctx, int(limit), int(offset), orderBy, sort, onlyEnabled); err != nil {
		return
	}

	list = make([]*m.Task, len(dbList))
	for i, dbVer := range dbList {
		list[i] = n.fromDb(dbVer)
	}

	return
}

func (n *Task) fromDb(dbVer *db.Task) (ver *m.Task) {
	ver = &m.Task{
		Id:          dbVer.Id,
		Name:        dbVer.Name,
		Description: dbVer.Description,
		Enabled:     dbVer.Enabled,
		AreaId:      dbVer.AreaId,
		Condition:   dbVer.Condition,
		CreatedAt:   dbVer.CreatedAt,
		UpdatedAt:   dbVer.UpdatedAt,
	}

	// triggers
	triggerAdaptor := GetTriggerAdaptor(n.db, n.orm)
	for _, dbVer := range dbVer.Triggers {
		tr := triggerAdaptor.fromDb(dbVer)
		ver.Triggers = append(ver.Triggers, tr)
	}

	// conditions
	conditionAdaptor := GetConditionAdaptor(n.db)
	for _, dbVer := range dbVer.Conditions {
		cond := conditionAdaptor.fromDb(dbVer)
		ver.Conditions = append(ver.Conditions, cond)
	}

	// actions
	actionAdaptor := GetActionAdaptor(n.db, n.orm)
	for _, dbVer := range dbVer.Actions {
		act := actionAdaptor.fromDb(dbVer)
		ver.Actions = append(ver.Actions, act)
	}

	// area
	if dbVer.Area != nil {
		areaAdaptor := GetAreaAdaptor(n.db)
		ver.Area = areaAdaptor.fromDb(dbVer.Area)
	}

	return
}

func (n *Task) toDb(ver *m.Task) (dbVer *db.Task) {
	dbVer = &db.Task{
		Id:          ver.Id,
		Name:        ver.Name,
		Description: ver.Description,
		Enabled:     ver.Enabled,
		Condition:   ver.Condition,
		AreaId:      ver.AreaId,
		CreatedAt:   ver.CreatedAt,
		UpdatedAt:   ver.UpdatedAt,
	}
	if len(ver.Triggers) > 0 {
		for _, tr := range ver.Triggers {
			dbVer.Triggers = append(dbVer.Triggers, &db.Trigger{Id: tr.Id})
		}
	}
	if len(ver.Conditions) > 0 {
		for _, tr := range ver.Conditions {
			dbVer.Conditions = append(dbVer.Conditions, &db.Condition{Id: tr.Id})
		}
	}
	if len(ver.Actions) > 0 {
		for _, tr := range ver.Actions {
			dbVer.Actions = append(dbVer.Actions, &db.Action{Id: tr.Id})
		}
	}

	return
}
