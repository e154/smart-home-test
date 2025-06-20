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

package onvif

import (
	"fmt"

	"github.com/e154/smart-home/internal/plugins/media/server"
	"github.com/e154/smart-home/internal/system/supervisor"
	"github.com/e154/smart-home/pkg/common"
	"github.com/e154/smart-home/pkg/events"
	m "github.com/e154/smart-home/pkg/models"
	"github.com/e154/smart-home/pkg/plugins"
)

// Actor ...
type Actor struct {
	*supervisor.BaseActor
	client      *Client
	snapshotUri *string
}

// NewActor ...
func NewActor(entity *m.Entity,
	service plugins.Service) (actor *Actor) {

	actor = &Actor{
		BaseActor: supervisor.NewBaseActor(entity, service),
	}

	actor.client = NewClient(actor.eventHandler)

	clientBind := NewClientBind(actor.client)

	// Actions
	for _, a := range actor.Actions {
		if a.ScriptEngine != nil {
			a.ScriptEngine.PushStruct("Camera", clientBind)
		}
	}

	if actor.ScriptsEngine != nil {
		actor.ScriptsEngine.PushStruct("Camera", clientBind)
	}

	if actor.Attrs == nil {
		actor.Attrs = NewAttr()
	}

	if actor.Setts == nil {
		actor.Setts = NewSettings()
	}

	if actor.Actions == nil {
		actor.Actions = NewActions()
	}

	return actor
}

func (a *Actor) Destroy() {
	a.Service.EventBus().Publish("system/media", server.EventRemoveList{Name: a.Id.String()})
	go a.client.Shutdown()
}

// Spawn ...
func (a *Actor) Spawn() {
	a.client.Start(a.Setts[AttrUserName].String(),
		a.Setts[AttrPassword].Decrypt(),
		a.Setts[AttrAddress].String(),
		a.Setts[AttrOnvifPort].Int64(),
		a.Setts[AttrRequireAuthorization].Bool())
	a.BaseActor.Spawn()
}

// SetState ...
func (a *Actor) SetState(params plugins.EntityStateParams) error {

	a.SetActorState(params.NewState)
	a.DeserializeAttr(params.AttributeValues)
	a.SaveState(false, params.StorageSave)

	return nil
}

func (a *Actor) addAction(event events.EventCallEntityAction) {
	a.runAction(event)
}

func (a *Actor) runAction(msg events.EventCallEntityAction) {
	if action, ok := a.Actions[msg.ActionName]; ok {
		if action.ScriptEngine != nil && action.ScriptEngine.Engine() != nil {
			if _, err := action.ScriptEngine.Engine().AssertFunction(FuncEntityAction, a.Id, action.Name, msg.Args); err != nil {
				log.Error(fmt.Errorf("entity id: %s: %w", a.Id, err).Error())
			}
			return
		}
	}
	if a.ScriptsEngine != nil && a.ScriptsEngine.Engine() != nil {
		if _, err := a.ScriptsEngine.AssertFunction(FuncEntityAction, a.Id, msg.ActionName, msg.Args); err != nil {
			log.Error(fmt.Errorf("entity id: %s: %w", a.Id, err).Error())
		}
	}
}

func (a *Actor) eventHandler(msg interface{}) {
	switch v := msg.(type) {
	case *StreamList:
		go a.prepareStreamList(v)
	case *ConnectionStatus:
		go a.updateState(v)
	case *MotionAlarm:
		go a.prepareMotionAlarm(v)
	}
}

func (a *Actor) updateState(event *ConnectionStatus) {
	info := a.Info()
	var newStat = AttrOffline
	if event.Connected {
		newStat = AttrConnected
	}
	if info.State != nil && info.State.Name == newStat {
		return
	}
	a.SetState(plugins.EntityStateParams{
		NewState:    common.String(newStat),
		StorageSave: true,
	})
}

func (a *Actor) prepareMotionAlarm(event *MotionAlarm) {
	a.SetState(plugins.EntityStateParams{
		NewState: common.String(AttrConnected),
		AttributeValues: m.AttributeValue{
			AttrMotion:     event.State,
			AttrMotionTime: event.Time,
		},
		StorageSave: true,
	})
}

func (a *Actor) prepareStreamList(event *StreamList) {
	a.snapshotUri = event.SnapshotUri
	a.Service.EventBus().Publish("system/media", server.EventUpdateList{
		Name:     a.Id.String(),
		Channels: event.List,
	})
}

func (a *Actor) GetSnapshotUri() string {
	return common.StringValue(a.snapshotUri)
}
