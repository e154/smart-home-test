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

package cgminer

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/e154/smart-home/common"
	m "github.com/e154/smart-home/models"
	"github.com/e154/smart-home/system/entity_manager"
	"github.com/e154/smart-home/system/event_bus"
	"github.com/e154/smart-home/system/plugins"
)

var (
	log = common.MustGetLogger("plugins.cgminer")
)

var _ plugins.Plugable = (*plugin)(nil)

func init() {
	plugins.RegisterPlugin(Name, New)
}

type plugin struct {
	*plugins.Plugin
	actorsLock *sync.Mutex
	actors     map[common.EntityId]*Actor
}

// New ...
func New() plugins.Plugable {
	return &plugin{
		Plugin:     plugins.NewPlugin(),
		actorsLock: &sync.Mutex{},
		actors:     make(map[common.EntityId]*Actor),
	}
}

// Load ...
func (p *plugin) Load(service plugins.Service) (err error) {
	if err = p.Plugin.Load(service); err != nil {
		return
	}

	p.EventBus.Subscribe(event_bus.TopicEntities, p.eventHandler)

	return nil
}

// Unload ...
func (p *plugin) Unload() (err error) {
	if err = p.Plugin.Unload(); err != nil {
		return
	}

	p.EventBus.Unsubscribe(event_bus.TopicEntities, p.eventHandler)

	return nil
}

// Name ...
func (p *plugin) Name() string {
	return Name
}

func (p *plugin) eventHandler(topic string, msg interface{}) {

	switch v := msg.(type) {
	case event_bus.EventStateChanged:
	case event_bus.EventCallAction:
		actor, ok := p.actors[v.EntityId]
		if !ok {
			return
		}
		actor.addAction(v)
	}

	return
}

// AddOrUpdateActor ...
func (p *plugin) AddOrUpdateActor(entity *m.Entity) (err error) {
	p.actorsLock.Lock()
	defer p.actorsLock.Unlock()

	if _, ok := p.actors[entity.Id]; ok {
		p.actors[entity.Id].Update()
		return
	}

	var actor *Actor
	if actor, err = NewActor(entity, p.EntityManager, p.Adaptors, p.ScriptService, p.EventBus); err != nil {
		return
	}
	p.actors[entity.Id] = actor
	p.EntityManager.Spawn(p.actors[entity.Id].Spawn)

	return
}

// RemoveActor ...
func (p *plugin) RemoveActor(entityId common.EntityId) error {
	return p.removeEntity(entityId)
}

func (p *plugin) removeEntity(name common.EntityId) (err error) {
	p.actorsLock.Lock()
	defer p.actorsLock.Unlock()

	if _, ok := p.actors[name]; !ok {
		err = errors.Wrap(common.ErrNotFound, fmt.Sprintf("failed remove \"%s\"", name))
		return
	}

	delete(p.actors, name)

	return
}

// Type ...
func (p *plugin) Type() plugins.PluginType {
	return plugins.PluginInstallable
}

// Depends ...
func (p *plugin) Depends() []string {
	return nil
}

// Version ...
func (p *plugin) Version() string {
	return "0.0.1"
}

// Options ...
func (p *plugin) Options() m.PluginOptions {
	return m.PluginOptions{
		Triggers:           false,
		Actors:             true,
		ActorCustomAttrs:   true,
		ActorAttrs:         NewAttr(),
		ActorCustomActions: true,
		ActorActions:       entity_manager.ToEntityActionShort(NewActions()),
		ActorCustomStates:  true,
		ActorStates:        entity_manager.ToEntityStateShort(NewStates()),
		ActorCustomSetts:   true,
		ActorSetts:         NewSettings(),
		Setts:              nil,
	}
}