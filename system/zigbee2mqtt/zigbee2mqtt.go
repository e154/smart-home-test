// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2020, Filippov Alex
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
	"github.com/e154/smart-home/adaptors"
	"github.com/e154/smart-home/system/graceful_service"
	"github.com/e154/smart-home/system/mqtt"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("zigbee2mqtt")
)

type Zigbee2mqtt struct {
	graceful  *graceful_service.GracefulService
	mqtt      *mqtt.Mqtt
	adaptors  *adaptors.Adaptors
	bridge    *Bridge
	isStarted bool
}

func NewZigbee2mqtt(graceful *graceful_service.GracefulService,
	mqtt *mqtt.Mqtt,
	adaptors *adaptors.Adaptors) *Zigbee2mqtt {
	return &Zigbee2mqtt{
		graceful: graceful,
		mqtt:     mqtt,
		bridge:   NewBridge(mqtt, adaptors),
		adaptors: adaptors,
	}
}

func (z *Zigbee2mqtt) Start() {
	if z.isStarted {
		return
	}
	z.bridge.Start()
	z.isStarted = true
}

func (z *Zigbee2mqtt) Shutdown() {
	if !z.isStarted {
		return
	}
	z.bridge.Stop()
	z.isStarted = false
}