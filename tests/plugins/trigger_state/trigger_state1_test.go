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

package trigger_state

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/e154/smart-home/internal/system/automation"
	"github.com/e154/smart-home/internal/system/migrations"
	"github.com/e154/smart-home/internal/system/zigbee2mqtt"
	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/common"
	"github.com/e154/smart-home/pkg/events"
	"github.com/e154/smart-home/pkg/models"
	"github.com/e154/smart-home/pkg/mqtt"
	"github.com/e154/smart-home/pkg/plugins"
	"github.com/e154/smart-home/pkg/scripts"

	"github.com/e154/bus"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"

	. "github.com/e154/smart-home/tests/plugins"
	. "github.com/e154/smart-home/tests/plugins/container"
)

func TestTriggerState1(t *testing.T) {

	const (
		zigbeeButtonId = "0x00158d00031c8ef3"

		buttonSourceScript = `# {"battery":100,"click":"long","duration":1515,"linkquality":126,"voltage":3042}
zigbee2mqttEvent =(message)->
  #print '---mqtt new event from button---'
  if !message
    return
  payload = unmarshal message.payload
  attrs =
    'battery': payload.battery
    'linkquality': payload.linkquality
    'voltage': payload.voltage
  state = ''
  if payload.action
    attrs.action = payload.action
    state = payload.action + "_action"
  if payload.click
    attrs.click = payload.click
    attrs.action = ""
    state = payload.click + "_click"
  EntitySetState ENTITY_ID,
    'new_state': state.toUpperCase()
    'attribute_values': attrs
    'storage_save': true
`

		task1SourceScript = `
automationTriggerStateChanged = (msg)->
  #print '---trigger---'
  p = msg.payload
  Done p.new_state.state.name
  return false
`
	)

	Convey("trigger state change", t, func(ctx C) {
		_ = BuildContainer().Invoke(func(adaptors *adaptors.Adaptors,
			scriptService scripts.ScriptService,
			supervisor plugins.Supervisor,
			zigbee2mqtt zigbee2mqtt.Zigbee2mqtt,
			mqttServer mqtt.MqttServ,
			automation automation.Automation,
			eventBus bus.Bus,
			migrations *migrations.Migrations) {

			migrations.Purge()

			// register plugins
			AddPlugin(adaptors, "triggers")
			AddPlugin(adaptors, "zigbee2mqtt")
			AddPlugin(adaptors, "state_change")

			// add zigbee2mqtt
			zigbeeServer := &models.Zigbee2mqtt{
				Name:       "main",
				Devices:    nil,
				PermitJoin: true,
				BaseTopic:  "zigbee2mqtt",
			}
			var err error
			zigbeeServer.Id, err = adaptors.Zigbee2mqtt.Add(context.Background(), zigbeeServer)
			So(err, ShouldBeNil)

			// add zigbee2mqtt_device
			buttonDevice := &models.Zigbee2mqttDevice{
				Id:            zigbeeButtonId,
				Zigbee2mqttId: zigbeeServer.Id,
				Name:          zigbeeButtonId,
				Model:         "WXKG01LM",
				Description:   "MiJia wireless switch",
				Manufacturer:  "Xiaomi",
				Status:        "active",
				Payload:       []byte("{}"),
			}
			err = adaptors.Zigbee2mqttDevice.Add(context.Background(), buttonDevice)
			So(err, ShouldBeNil)

			serviceCh := WaitService(eventBus, time.Second*5, "Mqtt", "Automation", "Zigbee2mqtt", "Supervisor")
			pluginsCh := WaitPlugins(eventBus, time.Second*5, "zigbee2mqtt", "triggers")

			mqttServer.Start()
			zigbee2mqtt.Start(context.Background())
			supervisor.Start(context.Background())
			automation.Start()

			defer mqttServer.Shutdown()
			defer zigbee2mqtt.Shutdown(context.Background())
			defer supervisor.Shutdown(context.Background())
			defer automation.Shutdown()

			So(<-serviceCh, ShouldBeTrue)
			So(<-pluginsCh, ShouldBeTrue)

			var counter atomic.Int32
			var lastStat atomic.String

			// common
			// ------------------------------------------------
			ch := make(chan struct{})
			scriptService.PushFunctions("Done", func(state string) {
				lastStat.Store(state)
				counter.Inc()
				if counter.Load() > 1 {
					close(ch)
				}
			})

			time.Sleep(time.Millisecond * 500)

			// add scripts
			// ------------------------------------------------

			buttonScript, err := AddScript("button", buttonSourceScript, adaptors, scriptService)
			So(err, ShouldBeNil)

			task1Script, err := AddScript("task", task1SourceScript, adaptors, scriptService)
			So(err, ShouldBeNil)

			// add entity
			// ------------------------------------------------
			buttonEnt := GetNewButton(fmt.Sprintf("zigbee2mqtt.%s", zigbeeButtonId), []*models.Script{buttonScript})
			err = adaptors.Entity.Add(context.Background(), buttonEnt)
			So(err, ShouldBeNil)

			eventBus.Publish("system/models/entities/"+buttonEnt.Id.String(), events.EventCreatedEntityModel{
				EntityId: buttonEnt.Id,
			})

			entityCh := WaitEntity(eventBus, time.Second*5, buttonEnt.Id.String())
			// wait entity
			So(<-entityCh, ShouldBeTrue)

			// automation
			// ------------------------------------------------
			trigger := &models.NewTrigger{
				Enabled:    true,
				Name:       "state_change",
				EntityIds:  []string{buttonEnt.Id.String()},
				ScriptId:   common.Int64(task1Script.Id),
				PluginName: "state_change",
			}
			triggerId, err := AddTrigger(trigger, adaptors, eventBus)
			So(err, ShouldBeNil)

			time.Sleep(time.Millisecond * 500)

			//TASK1
			newTask := &models.NewTask{
				Name:       "Toggle plug ON",
				Enabled:    true,
				TriggerIds: []int64{triggerId},
				Condition:  common.ConditionAnd,
			}
			id, err := AddTask(newTask, adaptors, eventBus)
			So(err, ShouldBeNil)

			taskCh := WaitTask(eventBus, time.Second*5, id)
			// wait task
			So(<-taskCh, ShouldBeTrue)

			// ------------------------------------------------

			mqttCli := mqttServer.NewClient("cli2")
			err = mqttCli.Publish("zigbee2mqtt/"+zigbeeButtonId, []byte(`{"battery":100,"action":"double","linkquality":134,"voltage":3042}`))
			So(err, ShouldBeNil)
			time.Sleep(time.Millisecond * 500)
			err = mqttCli.Publish("zigbee2mqtt/"+zigbeeButtonId, []byte(`{"battery":100,"click":"double","linkquality":134,"voltage":3042}`))
			So(err, ShouldBeNil)
			time.Sleep(time.Millisecond * 500)

			// wait message
			_, ok := WaitT[struct{}](time.Second*2, ch)
			ctx.So(ok, ShouldBeTrue)

			time.Sleep(time.Second)

			So(counter.Load(), ShouldBeGreaterThanOrEqualTo, 1)
			So(lastStat.Load(), ShouldEqual, "DOUBLE_CLICK")
		})
	})
}
