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

package mqtt_client

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	"github.com/e154/smart-home/pkg/logger"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	log = logger.MustGetLogger("mqtt_client")
)

// Client ...
type Client struct {
	cfg        *Config
	mx         sync.Mutex
	client     MQTT.Client
	subscribes map[string]Subscribe
}

// NewClient ...
func NewClient(cfg *Config) (client *Client, err error) {

	log.Infof("new queue client(%s) uri(%s)", cfg.ClientID, cfg.Broker)

	client = &Client{
		cfg:        cfg,
		subscribes: make(map[string]Subscribe),
	}

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetPingTimeout(time.Duration(cfg.PingTimeout) * time.Second).
		SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second).
		SetCleanSession(cfg.CleanSession).
		SetOnConnectHandler(client.onConnect).
		SetConnectionLostHandler(client.onConnectionLostHandler)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}

	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	client.client = MQTT.NewClient(opts)

	return
}

// Connect ...
func (c *Client) Connect() (err error) {

	c.mx.Lock()
	defer c.mx.Unlock()

	log.Infof("Connect to server %s", c.cfg.Broker)

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error().Error())
		err = token.Error()
	}

	return
}

// Disconnect ...
func (c *Client) Disconnect() {

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.client == nil {
		return
	}

	c.unsubscribeAll()

	c.mx.Lock()
	c.client.Disconnect(250)
	//c.client = nil
}

// Subscribe ...
func (c *Client) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) (err error) {

	if topic == "" {
		err = fmt.Errorf("%s: %w", "zero topic", apperr.ErrInternal)
		return
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	if _, ok := c.subscribes[topic]; !ok {
		c.subscribes[topic] = Subscribe{
			Qos:      qos,
			Callback: callback,
		}
	} else {
		err = fmt.Errorf("%s: %w", fmt.Sprintf("topic %s exist", topic), apperr.ErrInternal)
		return
	}

	if token := c.client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		err = token.Error()
	}
	return
}

// Unsubscribe ...
func (c *Client) Unsubscribe(topic string) (err error) {

	c.mx.Lock()
	defer c.mx.Unlock()

	if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		log.Error(token.Error().Error())
		return token.Error()
	}
	return
}

// unsubscribeAll ...
func (c *Client) unsubscribeAll() {
	c.mx.Lock()
	defer c.mx.Unlock()

	for topic := range c.subscribes {
		if token := c.client.Unsubscribe(topic); token.Error() != nil {
			log.Error(token.Error().Error())
		}
		delete(c.subscribes, topic)
	}
}

// Publish ...
func (c *Client) Publish(topic string, payload interface{}) (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.client != nil && (c.client.IsConnected()) {
		c.client.Publish(topic, c.cfg.Qos, false, payload)
	}
	return
}

// IsConnected ...
func (c *Client) IsConnected() bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c.client.IsConnectionOpen()
}

func (c *Client) onConnectionLostHandler(client MQTT.Client, e error) {

	c.mx.Lock()
	defer c.mx.Unlock()

	log.Debug("connection lost...")

	for topic := range c.subscribes {
		if token := c.client.Unsubscribe(topic); token.Error() != nil {
			log.Error(token.Error().Error())
		}
	}
}

func (c *Client) onConnect(client MQTT.Client) {

	c.mx.Lock()
	defer c.mx.Unlock()

	log.Debug("connected...")

	for topic, subscribe := range c.subscribes {
		if token := c.client.Subscribe(topic, subscribe.Qos, subscribe.Callback); token.Wait() && token.Error() != nil {
			log.Error(token.Error().Error())
		}
	}
}

// ClientIdGen ...
func ClientIdGen(args ...interface{}) string {
	var b strings.Builder
	b.WriteString("smarthome")
	for _, n := range args {
		fmt.Fprintf(&b, "_%v", n)
	}
	return b.String()
}
