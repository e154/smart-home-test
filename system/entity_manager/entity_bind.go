// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2021, Filippov Alex
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

package entity_manager

import (
	"github.com/e154/smart-home/common"
	m "github.com/e154/smart-home/models"
)

// Javascript Binding
//
// Entity
//  .setState(stateName)
//  .setAttributes(attrs)
//  .getAttributes()
//  .setMetric(name, value)
//  .callAction(name, value)
//

// EntityBind ...
type EntityBind struct {
	m.EntityShort
	manager EntityManager
}

// NewEntityBind ...
func NewEntityBind(id common.EntityId, manager EntityManager) *EntityBind {
	entity, _ := manager.GetEntityById(id)
	return &EntityBind{
		EntityShort: entity,
		manager:     manager,
	}
}

// SetState ...
func (e *EntityBind) SetState(stateName string) {
	e.manager.SetState(e.Id, EntityStateParams{
		NewState: common.String(stateName),
	})
}

// SetAttributes ...
func (e *EntityBind) SetAttributes(params m.AttributeValue) {
	e.manager.SetState(e.Id, EntityStateParams{
		AttributeValues: params,
	})
}

// GetAttributes ...
func (e *EntityBind) GetAttributes() m.AttributeValue {

	entity, err := e.manager.GetEntityById(e.Id)
	if err != nil {
		log.Error(err.Error())
	}

	return entity.Attributes.Serialize()
}

// SetMetric ...
func (e *EntityBind) SetMetric(name string, value map[string]interface{}) {
	e.manager.SetMetric(e.Id, name, value)
}

// CallAction ...
func (e *EntityBind) CallAction(action string, arg map[string]interface{}) {
	e.manager.CallAction(e.Id, action, arg)
}