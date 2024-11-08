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

package uptime

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/e154/smart-home/internal/system/supervisor"
	"github.com/e154/smart-home/pkg/common"
	"github.com/e154/smart-home/pkg/logger"
	m "github.com/e154/smart-home/pkg/models"
	"github.com/e154/smart-home/pkg/plugins"
)

const (
	name = "uptime"
)

var (
	log = logger.MustGetLogger("plugins.uptime")
)

var _ plugins.Pluggable = (*plugin)(nil)

//go:embed Readme.md
//go:embed Readme.ru.md
var F embed.FS

func init() {
	supervisor.RegisterPlugin(Name, New)
}

type plugin struct {
	*plugins.Plugin
	ticker     *time.Ticker
	storyModel *m.RunStory
}

// New ...
func New() plugins.Pluggable {
	p := &plugin{
		Plugin: plugins.NewPlugin(),
	}
	p.F = F
	return p
}

// Load ...
func (p *plugin) Load(ctx context.Context, service plugins.Service) (err error) {
	if err = p.Plugin.Load(ctx, service, p.ActorConstructor); err != nil {
		return
	}

	p.storyModel = &m.RunStory{
		Start: time.Now(),
	}

	p.storyModel.Id, err = p.Service.Adaptors().RunHistory.Add(context.Background(), p.storyModel)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if _, err = p.Service.Adaptors().Entity.GetById(context.Background(), common.EntityId(fmt.Sprintf("%s.%s", EntitySensor, Name))); err != nil {
		entity := &m.Entity{
			Id:         common.EntityId(fmt.Sprintf("%s.%s", EntitySensor, Name)),
			PluginName: Name,
			Attributes: NewAttr(),
		}
		err = p.Service.Adaptors().Entity.Add(context.Background(), entity)
	}

	go func() {
		const pause = 60
		p.ticker = time.NewTicker(time.Second * pause)

		for range p.ticker.C {

			if p.storyModel != nil {
				p.storyModel.End = common.Time(time.Now())
				if err = p.Service.Adaptors().RunHistory.Update(context.Background(), p.storyModel); err != nil {
					log.Error(err.Error())
				}
			}

			p.Actors.Range(func(key, value any) bool {
				actor, _ := value.(*Actor)
				actor.update()
				return true
			})
		}
	}()
	return nil
}

// Unload ...
func (p *plugin) Unload(ctx context.Context) (err error) {
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
	if err = p.Plugin.Unload(ctx); err != nil {
		return
	}

	if p.storyModel == nil {
		return
	}
	p.storyModel.End = common.Time(time.Now())
	if err = p.Service.Adaptors().RunHistory.Update(context.Background(), p.storyModel); err != nil {
		log.Error(err.Error())
	}
	return
}

// ActorConstructor ...
func (p *plugin) ActorConstructor(entity *m.Entity) (actor plugins.PluginActor, err error) {
	actor = NewActor(entity, p.Service)
	return
}

// Name ...
func (p plugin) Name() string {
	return name
}

// Depends ...
func (p *plugin) Depends() []string {
	return nil
}

// Options ...
func (p *plugin) Options() m.PluginOptions {
	return m.PluginOptions{
		Actors:             false,
		ActorCustomAttrs:   false,
		ActorAttrs:         NewAttr(),
		ActorCustomActions: false,
	}
}
