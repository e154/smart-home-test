// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2023, Filippov Alex
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

package automation

import (
	"context"
	"fmt"
	"sync"

	"github.com/e154/smart-home/internal/system/validation"
	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/apperr"
	"github.com/e154/smart-home/pkg/events"
	m "github.com/e154/smart-home/pkg/models"
	"github.com/e154/smart-home/pkg/plugins"
	"github.com/e154/smart-home/pkg/plugins/triggers"
	"github.com/e154/smart-home/pkg/scripts"

	"github.com/e154/bus"
	"go.uber.org/atomic"
)

type triggerManager struct {
	eventBus       bus.Bus
	scriptService  scripts.ScriptService
	supervisor     plugins.Supervisor
	adaptors       *adaptors.Adaptors
	isStarted      *atomic.Bool
	rawPlugin      triggers.IGetTrigger
	triggerCounter *atomic.Uint64
	validation     *validation.Validate
	sync.Mutex
	triggers map[int64]*Trigger
}

func NewTriggerManager(eventBus bus.Bus,
	scriptService scripts.ScriptService,
	sup plugins.Supervisor,
	adaptors *adaptors.Adaptors,
	validation *validation.Validate) (manager *triggerManager) {
	manager = &triggerManager{
		eventBus:       eventBus,
		scriptService:  scriptService,
		supervisor:     sup,
		adaptors:       adaptors,
		isStarted:      atomic.NewBool(false),
		triggers:       make(map[int64]*Trigger),
		triggerCounter: atomic.NewUint64(0),
		validation:     validation,
	}
	return
}

// Start ...
func (a *triggerManager) Start() {

	a.load()
	_ = a.eventBus.Subscribe("system/automation/triggers/+", a.eventHandler, false)
	_ = a.eventBus.Subscribe("system/models/triggers/+", a.eventHandler, false)
	a.isStarted.Store(true)

	log.Info("Started")
}

// Shutdown ...
func (a *triggerManager) Shutdown() {

	a.unload()
	_ = a.eventBus.Unsubscribe("system/automation/triggers/+", a.eventHandler)
	_ = a.eventBus.Unsubscribe("system/models/triggers/+", a.eventHandler)

	log.Info("Shutdown")
}

func (a *triggerManager) eventHandler(_ string, msg interface{}) {

	switch v := msg.(type) {
	case events.CommandEnableTrigger:
		go a.updateTrigger(v.Id)
	case events.CommandDisableTrigger:
		go a.removeTrigger(v.Id)

	case events.EventUpdatedTriggerModel:
		go a.updateTrigger(v.Id)
	case events.EventCreatedTriggerModel:
		go a.updateTrigger(v.Id)
	case events.EventRemovedTriggerModel:
		go a.removeTrigger(v.Id)
	}
}

func (a *triggerManager) load() {
	if a.isStarted.Load() {
		return
	}

	// load triggers plugin
	plugin, err := a.supervisor.GetPlugin(triggers.Name)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if rawPlugin, ok := plugin.(triggers.IGetTrigger); ok {
		a.rawPlugin = rawPlugin
	} else {
		log.Fatal("bad static cast triggers.IGetTrigger")
	}

	const perPage int64 = 500
	var page int64 = 0
LOOP:
	triggers, _, err := a.adaptors.Trigger.List(context.Background(), perPage, page*perPage, "", "", true)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, trigger := range triggers {
		if err = a.addTrigger(trigger); err != nil {
			log.Warn(err.Error())
		}
	}
	if len(triggers) != 0 {
		page++
		goto LOOP
	}

	a.bindScripts()

	log.Info("Loaded ...")
}

func (a *triggerManager) unload() {
	if !a.isStarted.Load() {
		return
	}

	for id := range a.triggers {
		a.removeTrigger(id)
	}
	a.isStarted.Store(false)

	log.Info("Unloaded ...")
}

// addTrigger ...
func (a *triggerManager) addTrigger(model *m.Trigger) (err error) {

	defer func() {
		if err == nil {
			a.triggerCounter.Inc()
		}
	}()

	if _, ok := a.triggers[model.Id]; ok {
		err = fmt.Errorf("trigger %s exist: %w", model.Name, apperr.ErrInternal)
		return
	}

	if !model.Enabled {
		return
	}

	var trigger *Trigger
	if trigger, err = NewTrigger(a.eventBus, a.scriptService, model, a.rawPlugin); err != nil {
		return
	}

	a.Lock()
	a.triggers[model.Id] = trigger
	a.Unlock()

	trigger.Start()

	return
}

// removeTrigger ...
func (a *triggerManager) removeTrigger(id int64) {
	a.Lock()
	defer a.Unlock()
	//log.Infof("remove trigger id:%d", id)

	trigger, ok := a.triggers[id]
	if !ok {
		return
	}
	trigger.Stop()
	delete(a.triggers, id)

	a.triggerCounter.Dec()
}

// updateTrigger ...
func (a *triggerManager) updateTrigger(id int64) {
	//log.Infof("reload trigger id:%d", id)
	a.removeTrigger(id)

	trigger, err := a.adaptors.Trigger.GetById(context.Background(), id)
	if err != nil {
		return
	}

	a.addTrigger(trigger)
}

func (a *triggerManager) IsLoaded(id int64) (loaded bool) {
	a.Lock()
	_, loaded = a.triggers[id]
	a.Unlock()
	return
}

func (a *triggerManager) bindScripts() {
	a.scriptService.PushFunctions("TriggerAdd", TriggerAdd(a))
	a.scriptService.PushFunctions("TriggerUpdate", TriggerUpdate(a))
	a.scriptService.PushFunctions("TriggerDelete", TriggerDelete(a))
}
