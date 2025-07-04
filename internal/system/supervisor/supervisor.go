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

package supervisor

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/e154/smart-home/internal/common"
	"github.com/e154/smart-home/internal/system/cache"
	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/apperr"
	pkgCommon "github.com/e154/smart-home/pkg/common"
	"github.com/e154/smart-home/pkg/events"
	"github.com/e154/smart-home/pkg/logger"
	"github.com/e154/smart-home/pkg/models"
	mqttTypes "github.com/e154/smart-home/pkg/mqtt"
	"github.com/e154/smart-home/pkg/plugins"
	"github.com/e154/smart-home/pkg/scheduler"
	"github.com/e154/smart-home/pkg/scripts"
	"github.com/e154/smart-home/pkg/web"

	"go.uber.org/atomic"
	"go.uber.org/fx"

	"github.com/e154/bus"
)

var (
	log = logger.MustGetLogger("supervisor")
)

type supervisor struct {
	*pluginManager
	scriptService     scripts.ScriptService
	eventScriptSubsMx sync.RWMutex
	eventScriptSubs   map[int64]map[pkgCommon.EntityId]struct{}
	cache             cache.Cache
}

// NewSupervisor ...
func NewSupervisor(lc fx.Lifecycle,
	adaptors *adaptors.Adaptors,
	bus bus.Bus,
	mqttServ mqttTypes.MqttServ,
	scriptService scripts.ScriptService,
	appConfig *models.AppConfig,
	eventBus bus.Bus,
	scheduler scheduler.Scheduler,
	crawler web.Crawler,
	authorization plugins.Authorization,
	httpAccessFilter plugins.HttpAccessFilter) plugins.Supervisor {
	s := &supervisor{
		scriptService:   scriptService,
		eventScriptSubs: make(map[int64]map[pkgCommon.EntityId]struct{}),
	}
	s.cache, _ = cache.NewCache("memory", `{"interval":60}`)
	s.pluginManager = &pluginManager{
		ExternalPlugins: NewExternalPlugins(adaptors),
		adaptors:        adaptors,
		isStarted:       atomic.NewBool(false),
		eventBus:        eventBus,
		enabledPlugins:  sync.Map{},
		pluginsWg:       &sync.WaitGroup{},
		service: &service{
			bus:              bus,
			supervisor:       s,
			mqttServ:         mqttServ,
			adaptors:         adaptors,
			scriptService:    scriptService,
			appConfig:        appConfig,
			scheduler:        scheduler,
			crawler:          crawler,
			authorization:    authorization,
			httpAccessFilter: httpAccessFilter,
		},
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return s.Shutdown(ctx)
		},
	})

	return s
}

func (e *supervisor) Start(ctx context.Context) (err error) {

	// event subscribe
	_ = e.eventBus.Subscribe("system/entities/+", e.eventHandler)
	_ = e.eventBus.Subscribe("system/models/entities/+", e.eventHandler)
	_ = e.eventBus.Subscribe("system/plugins/+", e.eventHandler)
	_ = e.eventBus.Subscribe("system/models/scripts/+", e.eventHandler)

	e.bindScripts()

	e.pluginManager.Start(ctx)

	_ = e.eventBus.Subscribe("system/services/scripts", e.handlerSystemScripts)
	e.eventBus.Publish("system/services/supervisor", events.EventServiceStarted{Service: "Supervisor"})

	log.Info("Started")

	return
}

// Shutdown ...
func (e *supervisor) Shutdown(ctx context.Context) (err error) {

	e.pluginManager.Shutdown(ctx)

	e.scriptService.PopFunction("GetEntity")
	e.scriptService.PopFunction("SetState")
	e.scriptService.PopFunction("SetStateName")
	e.scriptService.PopFunction("SetAttributes")
	e.scriptService.PopFunction("GetAttributes")
	e.scriptService.PopFunction("GetSettings")
	e.scriptService.PopFunction("SetMetric")
	e.scriptService.PopFunction("CallAction")
	e.scriptService.PopFunction("CallScene")
	e.scriptService.PopFunction("GetDistance")
	e.scriptService.PopFunction("PointInsideAria")

	_ = e.eventBus.Unsubscribe("system/services/scripts", e.handlerSystemScripts)
	_ = e.eventBus.Unsubscribe("system/entities/+", e.eventHandler)
	_ = e.eventBus.Unsubscribe("system/models/entities/+", e.eventHandler)
	_ = e.eventBus.Unsubscribe("system/plugins/+", e.eventHandler)
	_ = e.eventBus.Unsubscribe("system/models/scripts/+", e.eventHandler)

	e.eventBus.Publish("system/services/supervisor", events.EventServiceStopped{Service: "Supervisor"})

	log.Info("Shutdown")

	return
}

// Restart ...
func (e *supervisor) Restart(ctx context.Context) (err error) {
	if err = e.Shutdown(ctx); err != nil {
		return
	}
	err = e.Start(ctx)
	return
}

func (e *supervisor) handlerSystemScripts(_ string, event interface{}) {

	switch event.(type) {
	case events.EventServiceStarted, events.EventServiceRestarted:
		e.bindScripts()
	}
}

func (e *supervisor) bindScripts() {
	e.scriptService.PushFunctions("GetEntity", GetEntityBind(e))
	e.scriptService.PushFunctions("EntitySetState", SetStateBind(e))
	e.scriptService.PushFunctions("EntitySetStateName", SetStateNameBind(e))
	e.scriptService.PushFunctions("EntityGetState", GetStateBind(e))
	e.scriptService.PushFunctions("EntitySetAttributes", SetAttributesBind(e))
	e.scriptService.PushFunctions("EntityGetAttributes", GetAttributesBind(e))
	e.scriptService.PushFunctions("EntityGetSettings", GetSettingsBind(e))
	e.scriptService.PushFunctions("EntitySetMetric", SetMetricBind(e))
	e.scriptService.PushFunctions("EntityCallAction", CallActionBind(e))
	e.scriptService.PushFunctions("EntityCallScript", CallScriptBind(e))
	e.scriptService.PushFunctions("EntitiesCallAction", CallActionV2Bind(e))
	e.scriptService.PushFunctions("EntityCallScene", CallSceneBind(e))
	e.scriptService.PushFunctions("GeoDistanceToArea", GetDistanceToAreaBind(e.adaptors))
	e.scriptService.PushFunctions("GeoDistanceBetweenPoints", GetDistanceBetweenPointsBind(e.adaptors))
	e.scriptService.PushFunctions("GeoPointInsideArea", PointInsideAreaBind(e.adaptors))
	e.scriptService.PushFunctions("PushSystemEvent", PushSystemEvent(e))
}

// SetMetric ...
func (e *supervisor) SetMetric(id pkgCommon.EntityId, name string, value map[string]interface{}) {

	pla, err := e.GetActorById(id)
	if err != nil {
		return
	}

	pla.AddMetric(name, value)
}

// SetState ...
func (e *supervisor) SetState(id pkgCommon.EntityId, params plugins.EntityStateParams) (err error) {

	pla, err := e.GetActorById(id)
	if err != nil {
		return
	}

	if err = pla.SetState(params); err != nil {
		debug.PrintStack()
		log.Error(err.Error())
	}

	_ = e.cache.Delete(context.Background(), id.String())

	return
}

// EntityIsLoaded ...
func (e *supervisor) EntityIsLoaded(id pkgCommon.EntityId) (loaded bool) {

	if !e.PluginIsLoaded(id.PluginName()) {
		return
	}

	value, err := e.GetPlugin(id.PluginName())
	if err != nil {
		return
	}

	plugin := value.(plugins.Pluggable)
	loaded = plugin.EntityIsLoaded(id)

	return
}

// GetEntityById ...
func (e *supervisor) GetEntityById(id pkgCommon.EntityId) (entity models.EntityShort, err error) {

	var pla plugins.PluginActor
	if pla, err = e.GetActorById(id); err != nil {
		return
	}
	entity = NewEntity(pla)
	return
}

// GetActorById ...
func (e *supervisor) GetActorById(id pkgCommon.EntityId) (pla plugins.PluginActor, err error) {

	if !e.PluginIsLoaded(id.PluginName()) {
		err = fmt.Errorf("%s: %w", id.PluginName(), apperr.ErrPluginNotLoaded)
		return
	}

	var value interface{}
	if value, err = e.GetPlugin(id.PluginName()); err != nil {
		return
	}
	plugin := value.(plugins.Pluggable)

	pla, err = plugin.GetActor(id)

	return
}

// eventHandler ...
func (e *supervisor) eventHandler(_ string, message interface{}) {

	switch msg := message.(type) {
	case events.EventPluginLoaded:
		go func() { _ = e.eventLoadedPlugin(msg) }()
	case events.EventCreatedEntityModel:
		go e.eventCreatedEntity(msg)
	case events.EventUpdatedEntityModel:
		go e.eventUpdatedEntity(msg)
	case events.CommandUnloadEntity:
		go e.commandUnloadEntity(msg)
	case events.CommandLoadEntity:
		go e.commandLoadEntity(msg)
	case events.EventEntitySetState:
		go e.eventEntitySetState(msg)
	case events.EventGetLastState:
		go e.eventLastState(msg)
	case events.EventGetStateById:
		go e.eventGetStateById(msg)
	case events.EventUpdatedScriptModel:
		go e.eventUpdatedScript(msg)
	case events.EventRemovedScriptModel:
		go e.eventScriptDeleted(msg)
	case events.EventEntityLoaded:
		go e.eventEntityLoaded(msg)
	case events.EventEntityUnloaded:
		go e.eventEntityUnloaded(msg)
	}
}

func (e *supervisor) eventLastState(msg events.EventGetLastState) {

	if ok, _ := e.cache.IsExist(context.Background(), msg.EntityId.String()); ok {
		v, _ := e.cache.Get(context.Background(), msg.EntityId.String())
		state, ok := v.(events.EventLastStateChanged)
		if !ok {
			return
		}
		e.eventBus.Publish("system/entities/"+msg.EntityId.String(), state)
		return
	}
	_ = e.cache.Put(context.Background(), msg.EntityId.String(), nil, 10*time.Second)

	pla, err := e.GetActorById(msg.EntityId)
	if err != nil {
		return
	}

	info := pla.Info()

	state := events.EventLastStateChanged{
		PluginName: info.PluginName,
		EntityId:   info.Id,
	}

	if oldState := pla.GetOldState(); oldState != nil {
		state.OldState = *oldState
	}

	if newState := pla.GetCurrentState(); newState != nil {
		state.NewState = *newState
	}

	_ = e.cache.Put(context.Background(), msg.EntityId.String(), state, 30*time.Second)
	e.eventBus.Publish("system/entities/"+msg.EntityId.String(), state)
}

func (e *supervisor) eventGetStateById(msg events.EventGetStateById) {

	if msg.StorageId == 0 {
		return
	}

	list, err := e.adaptors.EntityStorage.GetLastThreeById(context.Background(), msg.EntityId, msg.StorageId)
	if err != nil {
		return
	}

	pla, err := e.GetActorById(msg.EntityId)
	if err != nil {
		return
	}

	info := pla.Info()

	state := events.EventStateById{
		UserID:     msg.Common.UserId(),
		SessionID:  msg.Common.SessionID,
		StorageId:  msg.StorageId,
		PluginName: info.PluginName,
		EntityId:   info.Id,
	}

	if len(list) > 0 {
		state.NewState = events.EventEntityState{
			EntityId:    info.Id,
			Value:       list[0].State,
			State:       nil,
			Attributes:  pla.Attributes().Copy(),
			Settings:    pla.Settings(),
			LastChanged: nil,
			LastUpdated: pkgCommon.Time(list[0].CreatedAt),
		}
		if newState, ok := info.States[list[0].State]; ok {
			state.NewState.State = &events.EntityState{
				Name:        newState.Name,
				Description: newState.Description,
				ImageUrl:    newState.ImageUrl,
				Icon:        newState.Icon,
			}
		}

		if _, err = state.NewState.Attributes.Deserialize(list[0].Attributes); err != nil {
			log.Error(err.Error())
		}
	}

	if len(list) > 1 {
		state.NewState.LastChanged = pkgCommon.Time(list[1].CreatedAt)
		state.OldState = events.EventEntityState{
			EntityId:    info.Id,
			Value:       list[1].State,
			State:       nil,
			Attributes:  pla.Attributes().Copy(),
			Settings:    pla.Settings(),
			LastChanged: nil,
			LastUpdated: pkgCommon.Time(list[1].CreatedAt),
		}
		if newState, ok := info.States[list[1].State]; ok {
			state.OldState.State = &events.EntityState{
				Name:        newState.Name,
				Description: newState.Description,
				ImageUrl:    newState.ImageUrl,
				Icon:        newState.Icon,
			}
		}

		if _, err = state.OldState.Attributes.Deserialize(list[1].Attributes); err != nil {
			log.Error(err.Error())
		}
	}

	if len(list) > 2 {
		state.OldState.LastChanged = pkgCommon.Time(list[2].CreatedAt)
	}

	e.eventBus.Publish("system/entities/"+msg.EntityId.String(), state)
}

func (e *supervisor) eventLoadedPlugin(msg events.EventPluginLoaded) (err error) {

	log.Infof("load plugin '%s' entities", msg.PluginName)

	var page int64
	var entities []*models.Entity
	const perPage = 500

LOOP:

	if entities, err = e.adaptors.Entity.GetByType(context.Background(), msg.PluginName, perPage, perPage*page); err != nil {
		log.Error(err.Error())
		return
	}

	for _, entity := range entities {
		go func(entity *models.Entity) {
			if err = e.AddEntity(entity); err != nil {
				log.Warnf("%s, %s", entity.Id, err.Error())
			}
		}(entity)
	}

	if len(entities) != 0 {
		page++
		goto LOOP
	}

	return
}

func (e *supervisor) eventCreatedEntity(msg events.EventCreatedEntityModel) {

	entity, err := e.adaptors.Entity.GetById(context.Background(), msg.EntityId)
	if err != nil {
		return
	}

	if !entity.AutoLoad {
		return
	}

	if err = e.AddEntity(entity); err != nil {
		log.Error(err.Error())
	}
}

func (e *supervisor) eventUpdatedEntity(msg events.EventUpdatedEntityModel) {
	e.updatedEntityById(msg.EntityId)
}

func (e *supervisor) updatedEntityById(entityId pkgCommon.EntityId) {
	entity, err := e.adaptors.Entity.GetById(context.Background(), entityId)
	if err != nil || !entity.AutoLoad {
		return
	}

	if err = e.UpdateEntity(entity); err != nil {
		log.Error(err.Error())
	}
}

func (e *supervisor) commandUnloadEntity(msg events.CommandUnloadEntity) {
	e.UnloadEntity(msg.EntityId)
}

func (e *supervisor) commandLoadEntity(msg events.CommandLoadEntity) {
	entity, err := e.adaptors.Entity.GetById(context.Background(), msg.EntityId)
	if err != nil {
		return
	}

	if !entity.AutoLoad {
		return
	}

	if err = e.AddEntity(entity); err != nil {
		log.Warnf("%s, %s", entity.Id, err.Error())
	}
}

func (e *supervisor) eventEntitySetState(msg events.EventEntitySetState) {

	_ = e.SetState(msg.EntityId, plugins.EntityStateParams{
		NewState:        msg.NewState,
		AttributeValues: msg.AttributeValues,
		SettingsValue:   msg.SettingsValue,
		StorageSave:     msg.StorageSave,
	})
}

// CallAction ...
func (e *supervisor) CallAction(id pkgCommon.EntityId, action string, arg map[string]interface{}) {
	e.eventBus.Publish("system/entities/"+id.String(), events.EventCallEntityAction{
		PluginName: pkgCommon.String(id.PluginName()),
		EntityId:   id.Ptr(),
		ActionName: action,
		Args:       arg,
	})
}

// CallScript ...
func (e *supervisor) CallScript(id pkgCommon.EntityId, fn string, arg ...interface{}) {
	pla, err := e.GetActorById(id)
	if err != nil {
		return
	}
	pla.CallScript(fn, arg...)
}

// CallActionV2 ...
func (e *supervisor) CallActionV2(params plugins.CallActionV2, arg map[string]interface{}) {
	var pluginName *string
	var entityId *pkgCommon.EntityId

	if params.EntityId != nil {
		entityId = params.EntityId.Ptr()
		pluginName = pkgCommon.String(entityId.PluginName())
	}

	e.eventBus.Publish("system/entities/", events.EventCallEntityAction{
		PluginName: pluginName,
		EntityId:   entityId,
		ActionName: params.ActionName,
		Args:       arg,
		AreaId:     params.AreaId,
		Tags:       params.Tags,
	})
}

// CallScene ...
func (e *supervisor) CallScene(id pkgCommon.EntityId, arg map[string]interface{}) {
	e.eventBus.Publish("system/entities/"+id.String(), events.EventCallScene{
		PluginName: id.PluginName(),
		EntityId:   id,
		Args:       arg,
	})
}

// AddEntity ...
func (e *supervisor) AddEntity(entity *models.Entity) (err error) {

	if !e.PluginIsLoaded(entity.PluginName) {
		err = fmt.Errorf("%s: %w", entity.PluginName, apperr.ErrPluginNotLoaded)
		return
	}

	var value interface{}
	if value, err = e.GetPlugin(entity.PluginName); err != nil {
		return
	}
	plugin := value.(plugins.Pluggable)
	err = plugin.AddOrUpdateActor(entity)
	return
}

// UpdateEntity ...
func (e *supervisor) UpdateEntity(entity *models.Entity) (err error) {

	if !e.PluginIsLoaded(entity.PluginName) {
		err = fmt.Errorf("%s: %w", entity.PluginName, apperr.ErrPluginNotLoaded)
		return
	}

	var value interface{}
	if value, err = e.GetPlugin(entity.PluginName); err != nil {
		return
	}

	plugin := value.(plugins.Pluggable)

	err = plugin.AddOrUpdateActor(entity)

	return
}

// UnloadEntity ...
func (e *supervisor) UnloadEntity(id pkgCommon.EntityId) {

	if !e.PluginIsLoaded(id.PluginName()) {
		return
	}

	value, err := e.GetPlugin(id.PluginName())
	if err != nil {
		return
	}

	plugin := value.(plugins.Pluggable)
	_ = plugin.RemoveActor(id)

	_ = e.cache.Delete(context.Background(), id.String())
}

func (e *supervisor) GetService() plugins.Service {
	return e.service
}

// watch to see if the scripts change
func (e *supervisor) eventUpdatedScript(msg events.EventUpdatedScriptModel) {

	if _, ok := e.eventScriptSubs[msg.ScriptId]; !ok {
		return
	}

	variable, err := e.adaptors.Variable.GetByName(context.Background(), "restartComponentIfScriptChanged")
	if err != nil || !variable.GetBool() {
		return
	}

	e.eventScriptSubsMx.RLock()
	defer e.eventScriptSubsMx.RUnlock()

	for entityId := range e.eventScriptSubs[msg.ScriptId] {
		go func(entityId pkgCommon.EntityId) {
			e.updatedEntityById(entityId)
		}(entityId)
	}
}

func (e *supervisor) eventScriptDeleted(msg events.EventRemovedScriptModel) {

	if _, ok := e.eventScriptSubs[msg.ScriptId]; !ok {
		return
	}

	variable, err := e.adaptors.Variable.GetByName(context.Background(), "restartComponentIfScriptChanged")
	if err != nil || !variable.GetBool() {
		return
	}

	e.eventScriptSubsMx.RLock()
	defer e.eventScriptSubsMx.RUnlock()

	for entityId := range e.eventScriptSubs[msg.ScriptId] {
		go func(entityId pkgCommon.EntityId) {
			e.UnloadEntity(entityId)
		}(entityId)
	}
}

func (e *supervisor) eventEntityLoaded(msg events.EventEntityLoaded) {
	go e.updateScriptWatcher(msg.EntityId)
}

func (e *supervisor) updateScriptWatcher(entityId pkgCommon.EntityId) {

	e.eventScriptSubsMx.Lock()
	defer e.eventScriptSubsMx.Unlock()

	entity, err := e.adaptors.Entity.GetById(context.Background(), entityId)
	if err != nil {
		return
	}

	if entity.Scripts != nil {
		for _, script := range entity.Scripts {
			if e.eventScriptSubs[script.Id] == nil {
				e.eventScriptSubs[script.Id] = make(map[pkgCommon.EntityId]struct{})
			}
			e.eventScriptSubs[script.Id][entity.Id] = struct{}{}
		}
	}

	if entity.Actions != nil {
		for _, action := range entity.Actions {
			if action.ScriptId != nil {
				if e.eventScriptSubs[*action.ScriptId] == nil {
					e.eventScriptSubs[*action.ScriptId] = make(map[pkgCommon.EntityId]struct{})
				}
				e.eventScriptSubs[*action.ScriptId][entity.Id] = struct{}{}
			}
		}
	}
}

func (e *supervisor) eventEntityUnloaded(msg events.EventEntityUnloaded) {

	e.eventScriptSubsMx.Lock()
	defer e.eventScriptSubsMx.Unlock()

	for scriptId := range e.eventScriptSubs {
		delete(e.eventScriptSubs[scriptId], msg.EntityId)
	}
}

//
// \watch to see if the scripts change
//

func (e *supervisor) PushSystemEvent(strCommand string, params map[string]interface{}) {

	var err error
	var topic string
	var command interface{}

	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered")
			debug.PrintStack()
		}
	}()

	switch strCommand {

	// tasks
	case "command_enable_task":
		cmd := events.CommandEnableTask{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/tasks/%d", cmd.Id)
		command = cmd
	case "command_disable_task":
		cmd := events.CommandDisableTask{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/tasks/%d", cmd.Id)
		command = cmd

	// triggers
	case "command_enable_trigger":
		cmd := events.CommandEnableTrigger{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/triggers/%d", cmd.Id)
		command = cmd
	case "command_disable_trigger":
		cmd := events.CommandDisableTrigger{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/triggers/%d", cmd.Id)
		command = cmd
	case "event_call_trigger":
		cmd := events.EventCallTrigger{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/triggers/%d", cmd.Id)
		command = cmd

	//actions
	case "event_call_action":
		cmd := events.EventCallAction{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/automation/actions/%d", cmd.Id)
		command = cmd

	// entity
	case "command_load_entity":
		cmd := events.CommandLoadEntity{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/models/entities/%s", cmd.EntityId)
		command = cmd
	case "command_unload_entity":
		cmd := events.CommandUnloadEntity{}
		err = common.SetFields(&cmd, params)
		topic = fmt.Sprintf("system/models/entities/%s", cmd.EntityId)
		command = cmd

	default:
		log.Warnf("unknown command %s", strCommand)
		return
	}

	if err != nil {
		log.Error(err.Error())
		return
	}

	e.eventBus.Publish(topic, command)
}
