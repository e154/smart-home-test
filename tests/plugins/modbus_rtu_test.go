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

package plugins

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/e154/smart-home/adaptors"
	"github.com/e154/smart-home/common"
	m "github.com/e154/smart-home/models"
	"github.com/e154/smart-home/plugins/modbus_rtu"
	"github.com/e154/smart-home/plugins/node"
	"github.com/e154/smart-home/system/automation"
	"github.com/e154/smart-home/system/entity_manager"
	"github.com/e154/smart-home/system/event_bus"
	"github.com/e154/smart-home/system/migrations"
	"github.com/e154/smart-home/system/mqtt"
	"github.com/e154/smart-home/system/scripts"
	. "github.com/smartystreets/goconvey/convey"
)

func TestModbusRtu(t *testing.T) {

	const plugActionOnOffSourceScript = `

getStatus =(status)->
    if status == 1
        return 'ON'
    else
        return 'OFF'

writeRegisters =(d, c, r)->
    return ModbusRtu 'WriteMultipleRegisters', d, c, r

checkStatus =->
    COMMAND = []
    FUNC = 'ReadHoldingRegisters'
    ADDRESS = 0
    COUNT = 16

    res = ModbusRtu FUNC, ADDRESS, COUNT, COMMAND
    So(res.error, 'ShouldEqual', '')
    So(res.result, 'ShouldEqual', '[1 0 1]')
    So(res.time, 'ShouldNotBeBlank', '')

doOnAction = ->
    res = writeRegisters(0, 1, [1])
    So(res.error, 'ShouldEqual', '')
    So(res.result, 'ShouldEqual', '[]')
    So(res.time, 'ShouldNotBeBlank', '')

doOnErrAction = ->
    res = writeRegisters(0, 1, [1])
    So(res.error, 'ShouldEqual', 'some error')
    So(res.result, 'ShouldEqual', '[]')
    So(res.time, 'ShouldNotBeBlank', '')

doOffAction = ->
    res = writeRegisters(0, 1, [0])
    So(res.error, 'ShouldEqual', '')
    So(res.result, 'ShouldEqual', '[]')
    So(res.time, 'ShouldNotBeBlank', '')

entityAction = (entityId, actionName)->
    #print '---action on/off--'
    switch actionName
        when 'ON' then doOnAction()
        when 'OFF' then doOffAction()
        when 'CHECK' then checkStatus()
        when 'ON_WITH_ERR' then doOnErrAction()
`

	Convey("modbus_rtu", t, func(ctx C) {
		_ = container.Invoke(func(adaptors *adaptors.Adaptors,
			migrations *migrations.Migrations,
			scriptService scripts.ScriptService,
			entityManager entity_manager.EntityManager,
			mqttServer mqtt.MqttServ,
			automation automation.Automation,
			eventBus event_bus.EventBus,
			pluginManager common.PluginManager) {

			eventBus.Purge()
			scriptService.Purge()

			err := migrations.Purge()
			So(err, ShouldBeNil)

			// bind convey
			RegisterConvey(scriptService, ctx)

			// register plugins
			err = AddPlugin(adaptors, "node")
			err = AddPlugin(adaptors, "triggers")
			err = AddPlugin(adaptors, "modbus_rtu")
			So(err, ShouldBeNil)

			go mqttServer.Start()

			// add scripts
			// ------------------------------------------------

			plugActionOnOffScript, err := AddScript("plug script", plugActionOnOffSourceScript, adaptors, scriptService)
			So(err, ShouldBeNil)

			// add entity
			// ------------------------------------------------
			nodeEnt := GetNewNode("main")
			err = adaptors.Entity.Add(nodeEnt)
			So(err, ShouldBeNil)

			plugEnt := GetNewModbusRtu("plug")
			plugEnt.ParentId = &nodeEnt.Id
			plugEnt.Actions = []*m.EntityAction{
				{
					Name:        "ON",
					Description: "включить",
					Script:      plugActionOnOffScript,
				},
				{
					Name:        "OFF",
					Description: "выключить",
					Script:      plugActionOnOffScript,
				},
				{
					Name:        "CHECK",
					Description: "condition check",
					Script:      plugActionOnOffScript,
				},
				{
					Name:        "ON_WITH_ERR",
					Description: "error case",
					Script:      plugActionOnOffScript,
				},
			}
			plugEnt.States = []*m.EntityState{
				{
					Name:        "ON",
					Description: "on state",
				},
				{
					Name:        "OFF",
					Description: "off state",
				},
				{
					Name:        "ERROR",
					Description: "error state",
				},
			}
			plugEnt.Settings[modbus_rtu.AttrSlaveId].Value = 1
			plugEnt.Settings[modbus_rtu.AttrBaud].Value = 19200
			plugEnt.Settings[modbus_rtu.AttrStopBits].Value = 1
			plugEnt.Settings[modbus_rtu.AttrTimeout].Value = 100
			plugEnt.Settings[modbus_rtu.AttrDataBits].Value = 8
			plugEnt.Settings[modbus_rtu.AttrParity].Value = "none"
			err = adaptors.Entity.Add(plugEnt)
			So(err, ShouldBeNil)
			_, err = adaptors.EntityStorage.Add(m.EntityStorage{
				EntityId:   plugEnt.Id,
				Attributes: plugEnt.Attributes.Serialize(),
			})
			So(err, ShouldBeNil)

			// ------------------------------------------------
			pluginManager.Start()
			automation.Reload()
			entityManager.SetPluginManager(pluginManager)
			entityManager.LoadEntities()

			defer func() {
				mqttServer.Shutdown()
				entityManager.Shutdown()
				automation.Shutdown()
				pluginManager.Shutdown()
			}()

			time.Sleep(time.Millisecond * 500)

			ch := make(chan []byte)
			mqttCli := mqttServer.NewClient("cli")
			mqttCli.Subscribe("home/node/main/req/#", func(cli mqtt.MqttCli, message mqtt.Message) {
				ch <- message.Payload
			})
			defer mqttCli.UnsubscribeAll()

			// commands
			t.Run("on command", func(t *testing.T) {
				entityManager.CallAction(plugEnt.Id, "ON", nil)

				ticker := time.NewTimer(time.Second * 2)
				defer ticker.Stop()

				var req []byte
				var ok bool
				select {
				case req = <-ch:
					ok = true
					break
				case <-ticker.C:
					break
				}

				ctx.So(ok, ShouldBeTrue)

				// what see node
				request := node.MessageRequest{}
				err = json.Unmarshal(req, &request)
				ctx.So(err, ShouldBeNil)

				cmd := modbus_rtu.ModBusCommand{}
				err = json.Unmarshal(request.Command, &cmd)
				ctx.So(err, ShouldBeNil)

				prop := map[string]interface{}{}
				err = json.Unmarshal(request.Properties, &prop)
				ctx.So(err, ShouldBeNil)

				ctx.So(request.EntityId, ShouldEqual, plugEnt.Id)
				ctx.So(request.DeviceType, ShouldEqual, modbus_rtu.DeviceTypeModbusRtu)
				ctx.So(prop["baud"], ShouldEqual, 19200)
				ctx.So(prop["data_bits"], ShouldEqual, 8)
				ctx.So(prop["parity"], ShouldEqual, "none")
				ctx.So(prop["slave_id"], ShouldEqual, 1)
				ctx.So(prop["sleep"], ShouldEqual, nil)
				ctx.So(prop["stop_bits"], ShouldEqual, 1)
				ctx.So(prop["timeout"], ShouldEqual, 100)

				ctx.So(cmd.Function, ShouldEqual, "WriteMultipleRegisters")
				ctx.So(cmd.Address, ShouldEqual, 0)
				ctx.So(cmd.Count, ShouldEqual, 1)
				ctx.So(cmd.Command, ShouldResemble, []uint16{1})

				// response from node
				r := modbus_rtu.ModBusResponse{
					Error:  "",
					Result: []uint16{},
				}
				b, _ := json.Marshal(r)
				resp := node.MessageResponse{
					EntityId:   plugEnt.Id,
					DeviceType: modbus_rtu.DeviceTypeModbusRtu,
					Properties: nil,
					Response:   b,
					Status:     "",
				}
				b, _ = json.Marshal(resp)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/plugin.test"), b)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/%s", plugEnt.Id), b)

				time.Sleep(time.Millisecond * 500)
			})

			t.Run("off command", func(t *testing.T) {
				entityManager.CallAction(plugEnt.Id, "OFF", nil)

				ticker := time.NewTimer(time.Second * 2)
				defer ticker.Stop()

				var req []byte
				var ok bool
				select {
				case req = <-ch:
					ok = true
					break
				case <-ticker.C:
					break
				}

				ctx.So(ok, ShouldBeTrue)

				// what see node
				request := node.MessageRequest{}
				err = json.Unmarshal(req, &request)
				ctx.So(err, ShouldBeNil)

				cmd := modbus_rtu.ModBusCommand{}
				err = json.Unmarshal(request.Command, &cmd)
				ctx.So(err, ShouldBeNil)

				prop := map[string]interface{}{}
				err = json.Unmarshal(request.Properties, &prop)
				ctx.So(err, ShouldBeNil)

				ctx.So(request.EntityId, ShouldEqual, plugEnt.Id)
				ctx.So(request.DeviceType, ShouldEqual, modbus_rtu.DeviceTypeModbusRtu)
				ctx.So(prop["baud"], ShouldEqual, 19200)
				ctx.So(prop["data_bits"], ShouldEqual, 8)
				ctx.So(prop["parity"], ShouldEqual, "none")
				ctx.So(prop["slave_id"], ShouldEqual, 1)
				ctx.So(prop["sleep"], ShouldEqual, nil)
				ctx.So(prop["stop_bits"], ShouldEqual, 1)
				ctx.So(prop["timeout"], ShouldEqual, 100)

				ctx.So(cmd.Function, ShouldEqual, "WriteMultipleRegisters")
				ctx.So(cmd.Address, ShouldEqual, 0)
				ctx.So(cmd.Count, ShouldEqual, 1)
				ctx.So(cmd.Command, ShouldResemble, []uint16{0})

				// response from node
				r := modbus_rtu.ModBusResponse{
					Error:  "",
					Result: []uint16{},
				}
				b, _ := json.Marshal(r)
				resp := node.MessageResponse{
					EntityId:   plugEnt.Id,
					DeviceType: modbus_rtu.DeviceTypeModbusRtu,
					Properties: nil,
					Response:   b,
					Status:     "",
				}
				b, _ = json.Marshal(resp)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/plugin.test"), b)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/%s", plugEnt.Id), b)

				time.Sleep(time.Millisecond * 500)
			})

			t.Run("check command", func(t *testing.T) {
				entityManager.CallAction(plugEnt.Id, "CHECK", nil)

				ticker := time.NewTimer(time.Second * 2)
				defer ticker.Stop()

				var req []byte
				var ok bool
				select {
				case req = <-ch:
					ok = true
					break
				case <-ticker.C:
					break
				}

				ctx.So(ok, ShouldBeTrue)

				// what see node
				request := node.MessageRequest{}
				err = json.Unmarshal(req, &request)
				ctx.So(err, ShouldBeNil)

				cmd := modbus_rtu.ModBusCommand{}
				err = json.Unmarshal(request.Command, &cmd)
				ctx.So(err, ShouldBeNil)

				prop := map[string]interface{}{}
				err = json.Unmarshal(request.Properties, &prop)
				ctx.So(err, ShouldBeNil)

				ctx.So(request.EntityId, ShouldEqual, plugEnt.Id)
				ctx.So(request.DeviceType, ShouldEqual, modbus_rtu.DeviceTypeModbusRtu)
				ctx.So(prop["baud"], ShouldEqual, 19200)
				ctx.So(prop["data_bits"], ShouldEqual, 8)
				ctx.So(prop["parity"], ShouldEqual, "none")
				ctx.So(prop["slave_id"], ShouldEqual, 1)
				ctx.So(prop["sleep"], ShouldEqual, nil)
				ctx.So(prop["stop_bits"], ShouldEqual, 1)
				ctx.So(prop["timeout"], ShouldEqual, 100)

				ctx.So(cmd.Function, ShouldEqual, "ReadHoldingRegisters")
				ctx.So(cmd.Address, ShouldEqual, 0)
				ctx.So(cmd.Count, ShouldEqual, 16)
				ctx.So(cmd.Command, ShouldResemble, []uint16{})

				// response from node
				r := modbus_rtu.ModBusResponse{
					Error:  "",
					Result: []uint16{1, 0, 1},
				}
				b, _ := json.Marshal(r)
				resp := node.MessageResponse{
					EntityId:   plugEnt.Id,
					DeviceType: modbus_rtu.DeviceTypeModbusRtu,
					Properties: nil,
					Response:   b,
					Status:     "",
				}
				b, _ = json.Marshal(resp)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/plugin.test"), b)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/%s", plugEnt.Id), b)

				time.Sleep(time.Millisecond * 500)
			})

			t.Run("bad command", func(t *testing.T) {
				entityManager.CallAction(plugEnt.Id, "NULL", nil)

				ticker := time.NewTimer(time.Second * 1)
				defer ticker.Stop()

				var ok bool
				select {
				case <-ch:
					ok = true
					break
				case <-ticker.C:
					break
				}

				ctx.So(ok, ShouldBeFalse)
			})

			t.Run("response with error", func(t *testing.T) {
				entityManager.CallAction(plugEnt.Id, "ON_WITH_ERR", nil)

				ticker := time.NewTimer(time.Second * 2)
				defer ticker.Stop()

				var ok bool
				select {
				case <-ch:
					ok = true
					break
				case <-ticker.C:
					break
				}

				ctx.So(ok, ShouldBeTrue)

				r := modbus_rtu.ModBusResponse{
					Error: "some error",
				}
				b, _ := json.Marshal(r)
				resp := node.MessageResponse{
					EntityId:   plugEnt.Id,
					DeviceType: modbus_rtu.DeviceTypeModbusRtu,
					Properties: nil,
					Response:   b,
					Status:     "",
				}
				b, _ = json.Marshal(resp)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/plugin.test"), b)
				mqttCli.Publish(fmt.Sprintf("home/node/main/resp/%s", plugEnt.Id), b)

				time.Sleep(time.Millisecond * 500)
			})
		})
	})
}