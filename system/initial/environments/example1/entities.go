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

package example1

import (
	"fmt"
	"os"

	"github.com/e154/smart-home/adaptors"
	"github.com/e154/smart-home/common"
	m "github.com/e154/smart-home/models"
	"github.com/e154/smart-home/plugins/cgminer"
	"github.com/e154/smart-home/plugins/cgminer/bitmine"
	"github.com/e154/smart-home/plugins/sensor"
	"github.com/e154/smart-home/plugins/telegram"
	. "github.com/e154/smart-home/system/initial/assertions"
)

// EntityManager ...
type EntityManager struct {
	adaptors *adaptors.Adaptors
}

// NewEntityManager ...
func NewEntityManager(adaptors *adaptors.Adaptors) *EntityManager {
	return &EntityManager{
		adaptors: adaptors,
	}
}

// Create ...
func (e *EntityManager) Create(scripts []*m.Script, areas []*m.Area) []*m.Entity {

	var script *m.Script
	if len(scripts) > 0 {
		script = scripts[0]
	}
	var area *m.Area
	if len(areas) > 0 {
		area = areas[0]
	}
	entity1 := e.addL3("l3n1", "192.168.0.247", script, area)
	entity2 := e.addL3("l3n2", "192.168.0.242", script, area)
	entity3 := e.addL3("l3n3", "192.168.0.244", script, area)
	entity4 := e.addL3("l3n4", "192.168.0.243", script, area)
	entity5 := e.addL3("l3n5", "192.168.0.240", script, area)

	tgBot := e.addTgBot("clavicus", os.Getenv("SH_TG_BOT_TOKEN"), script, area)
	sensorEntity := e.addSensor("api", scripts[1], area)

	return []*m.Entity{entity1, entity2, entity3, entity4, entity5, tgBot, sensorEntity}
}

func (e *EntityManager) addL3(name, host string, script *m.Script, area *m.Area) (ent *m.Entity) {
	settings := cgminer.NewSettings()
	settings[cgminer.SettingHost].Value = host
	settings[cgminer.SettingPort].Value = 4028
	settings[cgminer.SettingTimeout].Value = 2
	settings[cgminer.SettingUser].Value = "user"
	settings[cgminer.SettingPass].Value = "pass"
	settings[cgminer.SettingManufacturer].Value = bitmine.ManufactureBitmine
	settings[cgminer.SettingModel].Value = bitmine.DeviceL3Plus
	ent = &m.Entity{
		Id:          common.EntityId(fmt.Sprintf("cgminer.%s", name)),
		Description: "antminer L3+",
		PluginName:  cgminer.Name,
		AutoLoad:    true,
		Attributes:  cgminer.NewAttr(),
		Settings:    settings,
		Area:        area,
	}
	ent.Actions = []*m.EntityAction{
		{
			Name:        "CHECK",
			Description: "condition check",
			Script:      script,
		},
	}
	ent.States = []*m.EntityState{
		{
			Name:        "ENABLED",
			Description: "enabled state",
		},
		{
			Name:        "DISABLED",
			Description: "disabled state",
		},
		{
			Name:        "ERROR",
			Description: "error state",
		},
		{
			Name:        "WARNING",
			Description: "warning state",
		},
	}
	ent.Attributes = m.Attributes{
		"heat": {
			Name: "heat",
			Type: common.AttributeBool,
		},
		"chain1_temp_chip": {
			Name: "chain1_temp_chip",
			Type: common.AttributeInt,
		},
		"chain2_temp_chip": {
			Name: "chain2_temp_chip",
			Type: common.AttributeInt,
		},
		"chain3_temp_chip": {
			Name: "chain3_temp_chip",
			Type: common.AttributeInt,
		},
		"chain4_temp_chip": {
			Name: "chain4_temp_chip",
			Type: common.AttributeInt,
		},
		"chain1_temp_pcb": {
			Name: "chain1_temp_pcb",
			Type: common.AttributeInt,
		},
		"chain2_temp_pcb": {
			Name: "chain2_temp_pcb",
			Type: common.AttributeInt,
		},
		"chain3_temp_pcb": {
			Name: "chain3_temp_pcb",
			Type: common.AttributeInt,
		},
		"chain4_temp_pcb": {
			Name: "chain4_temp_pcb",
			Type: common.AttributeInt,
		},
		"chain_acn1": {
			Name: "chain_acn1",
			Type: common.AttributeInt,
		},
		"chain_acn2": {
			Name: "chain_acn2",
			Type: common.AttributeInt,
		},
		"chain_acn3": {
			Name: "chain_acn3",
			Type: common.AttributeInt,
		},
		"chain_acn4": {
			Name: "chain_acn4",
			Type: common.AttributeInt,
		},
		"fan1": {
			Name: "fan1",
			Type: common.AttributeInt,
		},
		"fan2": {
			Name: "fan2",
			Type: common.AttributeInt,
		},
		"ghs_av": {
			Name: "ghs_av",
			Type: common.AttributeInt,
		},
		"hardware_errors": {
			Name: "hardware_errors",
			Type: common.AttributeInt,
		},
	}

	err := e.adaptors.Entity.Add(ent)
	So(err, ShouldBeNil)

	_, err = e.adaptors.EntityStorage.Add(m.EntityStorage{
		EntityId:   ent.Id,
		Attributes: ent.Attributes.Serialize(),
	})
	So(err, ShouldBeNil)

	return
}

func (e *EntityManager) addTgBot(name, token string, script *m.Script, area *m.Area) (ent *m.Entity) {

	settings := telegram.NewSettings()
	settings[telegram.AttrToken].Value = token
	ent = &m.Entity{
		Id:          common.EntityId(fmt.Sprintf("%s.%s", telegram.Name, name)),
		Description: "",
		PluginName:  telegram.Name,
		AutoLoad:    true,
		Attributes:  telegram.NewAttr(),
		Settings:    settings,
		Actions: []*m.EntityAction{
			{
				Name:        "CHECK",
				Description: "check status",
				Script:      script,
			},
		},
		Area: area,
	}
	err := e.adaptors.Entity.Add(ent)
	So(err, ShouldBeNil)
	_, err = e.adaptors.EntityStorage.Add(m.EntityStorage{
		EntityId:   ent.Id,
		Attributes: ent.Attributes.Serialize(),
	})
	So(err, ShouldBeNil)

	return
}

func (e *EntityManager) addSensor(name string, script *m.Script, area *m.Area) (ent *m.Entity) {

	ent = &m.Entity{
		Id:          common.EntityId(fmt.Sprintf("%s.%s", sensor.Name, name)),
		Description: "",
		PluginName:  sensor.Name,
		AutoLoad:    true,
		Attributes: m.Attributes{
			"paid_rewards": {
				Name: "paid_rewards",
				Type: common.AttributeFloat,
			},
		},
		Actions: []*m.EntityAction{
			{
				Name:        "CHECK",
				Description: "condition check",
				Script:      script,
			},
		},
		States: []*m.EntityState{
			{
				Name:        "ENABLED",
				Description: "enabled state",
			},
			{
				Name:        "DISABLED",
				Description: "disabled state",
			},
			{
				Name:        "ERROR",
				Description: "error state",
			},
		},
		Area: area,
	}
	err := e.adaptors.Entity.Add(ent)
	So(err, ShouldBeNil)
	_, err = e.adaptors.EntityStorage.Add(m.EntityStorage{
		EntityId:   ent.Id,
		Attributes: ent.Attributes.Serialize(),
	})
	So(err, ShouldBeNil)

	return
}

// Upgrade ...
func (e EntityManager) Upgrade(oldVersion int, scripts []*m.Script, areas []*m.Area) (entities []*m.Entity, err error) {

	switch oldVersion {
	case 3:
		entity5 := e.addL3("l3n5", "192.168.0.240", scripts[0], areas[0])
		entities = append(entities, entity5)
	default:
		return
	}
	return
}