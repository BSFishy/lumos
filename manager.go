package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var manager = &Manager{}

type Manager struct{}

func (m *Manager) Start(c mqtt.Client, friendlyName string, groupConfig GroupConfig) {
	cm := &ColorManager{
		client:       c,
		friendlyName: friendlyName,
		groupConfig:  groupConfig,
	}

	cm.Run()
}

type ColorManager struct {
	client mqtt.Client

	friendlyName string
	groupConfig  GroupConfig

	previousColor, nextColor Oklab
	start, end               time.Time
}

func (c *ColorManager) Run() {
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", c.friendlyName)
	c.previousColor = c.groupConfig.SelectColor()

outer:
	for {
		c.nextColor = c.groupConfig.SelectColor()

		duration := c.groupConfig.Transition.Select()

		c.start = time.Now()
		c.end = c.start.Add(duration)

		ticker := time.NewTicker(config.TimestepDuration())
		timer := time.NewTimer(duration)

		c.updateColor(topic, duration.Seconds())

		for {
			select {
			case <-ticker.C:
				c.updateColor(topic, duration.Seconds())

			case <-timer.C:
				ticker.Stop()

				publish(c.client, topic, 1, false, ColorPayload(c.nextColor, 0))
				c.previousColor = c.nextColor

				// TODO: hold

				continue outer
			}
		}
	}
}

func (c *ColorManager) updateColor(topic string, durationSeconds float64) {
	transition := min(max(time.Until(c.end), 0), config.TimestepDuration())

	elapsed := time.Since(c.start).Seconds()
	t := clamp01((elapsed + transition.Seconds()) / durationSeconds)

	colorStep := c.previousColor.Lerp(c.nextColor, t)
	publish(c.client, topic, 1, false, ColorPayload(colorStep, transition.Seconds()))
}
