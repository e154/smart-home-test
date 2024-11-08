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

package models

import (
	"time"

	"github.com/e154/smart-home/pkg/common"
	"github.com/e154/smart-home/pkg/common/telemetry"
)

// Task ...
type Task struct {
	Triggers    []*Trigger           `json:"triggers" validate:"dive"`
	Conditions  []*Condition         `json:"conditions" validate:"dive"`
	Actions     []*Action            `json:"actions" validate:"dive"`
	Telemetry   telemetry.Telemetry  `json:"telemetry"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Name        string               `json:"name" validate:"required,lte=255"`
	Description string               `json:"description" validate:"lte=255"`
	Condition   common.ConditionType `json:"condition" validate:"required,oneof=or and"`
	Id          int64                `json:"id"`
	Area        *Area                `json:"area"`
	AreaId      *int64               `json:"area_id"`
	Enabled     bool                 `json:"enabled"`
	IsLoaded    bool                 `json:"is_loaded"`
}

// AddTrigger ...
func (t *Task) AddTrigger(tr *Trigger) {
	t.Triggers = append(t.Triggers, tr)
}

// AddCondition ...
func (t *Task) AddCondition(c *Condition) {
	t.Conditions = append(t.Conditions, c)
}

// AddAction ...
func (t *Task) AddAction(a *Action) {
	t.Actions = append(t.Actions, a)
}

// NewTask ...
type NewTask struct {
	TriggerIds   []int64              `json:"triggers" validate:"dive"`
	ConditionIds []int64              `json:"conditions" validate:"dive"`
	ActionIds    []int64              `json:"actions" validate:"dive"`
	Name         string               `json:"name" validate:"required,lte=255"`
	Description  string               `json:"description" validate:"lte=255"`
	Condition    common.ConditionType `json:"condition" validate:"required,oneof=or and"`
	Area         *Area                `json:"area"`
	AreaId       *int64               `json:"area_id"`
	Enabled      bool                 `json:"enabled"`
	IsLoaded     bool                 `json:"is_loaded"`
}

// UpdateTask ...
type UpdateTask struct {
	TriggerIds   []int64 `json:"triggers" validate:"dive"`
	ConditionIds []int64 `json:"conditions" validate:"dive"`
	ActionIds    []int64 `json:"actions" validate:"dive"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string               `json:"name" validate:"required,lte=255"`
	Description  string               `json:"description" validate:"lte=255"`
	Condition    common.ConditionType `json:"condition" validate:"required,oneof=or and"`
	Id           int64                `json:"id"`
	Area         *Area                `json:"area"`
	AreaId       *int64               `json:"area_id"`
	Enabled      bool                 `json:"enabled"`
	IsLoaded     bool                 `json:"is_loaded"`
}
