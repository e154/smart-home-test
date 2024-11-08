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

package zigbee2mqtt

import (
	"context"
	"sync"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/e154/smart-home/pkg/common/encryptor"
	mqttTypes "github.com/e154/smart-home/pkg/mqtt"

	"github.com/e154/bus"
	"go.uber.org/atomic"
	"go.uber.org/fx"

	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/events"
	"github.com/e154/smart-home/pkg/logger"
	m "github.com/e154/smart-home/pkg/models"
)

var (
	log = logger.MustGetLogger("zigbee2mqtt")
)

// zigbee2mqtt ...
type zigbee2mqtt struct {
	mqtt        mqttTypes.MqttServ
	adaptors    *adaptors.Adaptors
	isStarted   *atomic.Bool
	bridgesLock *sync.Mutex
	bridges     map[int64]*Bridge
	eventBus    bus.Bus
}

// NewZigbee2mqtt ...
func NewZigbee2mqtt(lc fx.Lifecycle,
	mqtt mqttTypes.MqttServ,
	adaptors *adaptors.Adaptors,
	eventBus bus.Bus) Zigbee2mqtt {
	zigbee2mqtt := &zigbee2mqtt{
		mqtt:        mqtt,
		adaptors:    adaptors,
		bridgesLock: &sync.Mutex{},
		bridges:     make(map[int64]*Bridge),
		eventBus:    eventBus,
		isStarted:   atomic.NewBool(false),
	}

	lc.Append(fx.Hook{
		OnStart: zigbee2mqtt.Start,
		OnStop:  zigbee2mqtt.Shutdown,
	})
	return zigbee2mqtt
}

// Start ...
func (z *zigbee2mqtt) Start(ctx context.Context) (err error) {
	if z.isStarted.Load() {
		return
	}
	z.isStarted.Store(true)

	models, _, err := z.adaptors.Zigbee2mqtt.List(ctx, 99, 0)
	if err != nil {
		log.Error(err.Error())
	}

	if len(models) == 0 {
		model := &m.Zigbee2mqtt{
			Name:       "zigbee2mqtt",
			BaseTopic:  "zigbee2mqtt",
			PermitJoin: true,
		}
		model.Id, err = z.adaptors.Zigbee2mqtt.Add(ctx, model)
		if err != nil {
			log.Error(err.Error())
			return
		}
		models = append(models, model)
	}

	if err := z.mqtt.Authenticator().Register(z.Authenticator); err != nil {
		log.Error(err.Error())
	}

	//todo fix race condition
	for _, model := range models {
		bridge := NewBridge(z.mqtt, z.adaptors, model)
		bridge.Start()

		z.bridgesLock.Lock()
		z.bridges[model.Id] = bridge
		z.bridgesLock.Unlock()
	}

	_ = z.eventBus.Subscribe("system/models/zigbee2mqtt/+", z.eventHandler)

	z.eventBus.Publish("system/services/zigbee2mqtt", events.EventServiceStarted{Service: "Zigbee2mqtt"})

	return
}

// Shutdown ...
func (z *zigbee2mqtt) Shutdown(ctx context.Context) (err error) {
	if !z.isStarted.Load() {
		return
	}
	z.isStarted.Store(false)
	for _, bridge := range z.bridges {
		bridge.Stop(context.Background())
	}
	_ = z.mqtt.Authenticator().Unregister(z.Authenticator)

	_ = z.eventBus.Unsubscribe("system/models/zigbee2mqtt/+", z.eventHandler)

	z.eventBus.Publish("system/services/zigbee2mqtt", events.EventServiceStopped{Service: "Zigbee2mqtt"})

	return
}

// AddBridge ...
func (z *zigbee2mqtt) AddBridge(model *m.Zigbee2mqtt) (err error) {

	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	bridge := NewBridge(z.mqtt, z.adaptors, model)
	bridge.Start()
	z.bridges[model.Id] = bridge
	return
}

// GetBridgeInfo ...
func (z *zigbee2mqtt) GetBridgeInfo(id int64) (*m.Zigbee2mqttInfo, error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	if br, ok := z.bridges[id]; ok {
		return br.Info(), nil
	}
	return nil, apperr.ErrNotFound
}

// UpdateBridge ...
func (z *zigbee2mqtt) UpdateBridge(model *m.Zigbee2mqtt) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(model.Id); err == nil {
		bridge.Stop(context.Background())
		delete(z.bridges, model.Id)
	} else {
		return
	}

	bridge = NewBridge(z.mqtt, z.adaptors, model)
	bridge.Start()
	z.bridges[model.Id] = bridge

	return
}

// DeleteBridge ...
func (z *zigbee2mqtt) DeleteBridge(bridgeId int64) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		bridge.Stop(context.Background())
		delete(z.bridges, bridgeId)
	}

	return
}

// ResetBridge ...
func (z *zigbee2mqtt) ResetBridge(bridgeId int64) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		bridge.ConfigReset()
	}
	return
}

// BridgeDeviceBan ...
func (z *zigbee2mqtt) BridgeDeviceBan(bridgeId int64, friendlyName string) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		bridge.Ban(friendlyName)
	}
	return
}

// BridgeDeviceWhitelist ...
func (z *zigbee2mqtt) BridgeDeviceWhitelist(bridgeId int64, friendlyName string) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		bridge.Whitelist(friendlyName)
	}
	return
}

// BridgeNetworkmap ...
func (z *zigbee2mqtt) BridgeNetworkmap(bridgeId int64) (networkmap string, err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		networkmap = bridge.Networkmap()
	}
	return
}

// BridgeUpdateNetworkmap ...
func (z *zigbee2mqtt) BridgeUpdateNetworkmap(bridgeId int64) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	var bridge *Bridge
	if bridge, err = z.unsafeGetBridge(bridgeId); err == nil {
		bridge.UpdateNetworkmap()
	}
	return
}

func (z *zigbee2mqtt) unsafeGetBridge(bridgeId int64) (bridge *Bridge, err error) {
	var ok bool
	if bridge, ok = z.bridges[bridgeId]; !ok {
		err = apperr.ErrNotFound
	}
	return
}

// GetTopicByDevice ...
func (z *zigbee2mqtt) GetTopicByDevice(model *m.Zigbee2mqttDevice) (topic string, err error) {

	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	br, ok := z.bridges[model.Zigbee2mqttId]
	if !ok {
		err = apperr.ErrNotFound
		return
	}

	topic = br.GetDeviceTopic(model.Id)

	return
}

// DeviceRename ...
func (z *zigbee2mqtt) DeviceRename(friendlyName, name string) (err error) {
	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	for _, bridge := range z.bridges {
		_ = bridge.RenameDevice(friendlyName, name)
	}

	return
}

// Authenticator ...
func (z *zigbee2mqtt) Authenticator(login, password string) (err error) {

	z.bridgesLock.Lock()
	defer z.bridgesLock.Unlock()

	for _, bridge := range z.bridges {
		if bridge.model.Login != login {
			err = apperr.ErrBadLoginOrPassword
			return
		}

		if ok := encryptor.CheckPasswordHash(password, bridge.model.EncryptedPassword); ok {
			err = nil
			return
		}
	}

	err = apperr.ErrBadLoginOrPassword

	return
}

// eventHandler ...
func (z *zigbee2mqtt) eventHandler(_ string, message interface{}) {

	var err error
	switch msg := message.(type) {
	case events.EventCreatedZigbee2mqttModel:
		err = z.AddBridge(msg.Bridge)
	case events.EventUpdatedZigbee2mqttModel:
		err = z.UpdateBridge(msg.Bridge)
	case events.EventRemovedZigbee2mqttModel:
		err = z.DeleteBridge(msg.Id)
	}

	if err != nil {
		log.Error(err.Error())
	}
}
