package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var manager = &Manager{}

type Manager struct{}

func (m *Manager) Start(c mqtt.Client, friendlyName string) {
	cm := &ColorManager{
		client:       c,
		friendlyName: friendlyName,
	}

	cm.Run()
}

type ColorManager struct {
	client mqtt.Client

	friendlyName string

	previousColor, nextColor HSB
	start, end               time.Time
}

func (c *ColorManager) Run() {
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", c.friendlyName)
	c.previousColor = chooseColor()

outer:
	for {
		c.nextColor = chooseColor()

		durationSeconds := config.speedMin.Seconds() + rand.Float64()*(config.speedMax.Seconds()-config.speedMin.Seconds())
		duration := time.Duration(durationSeconds * float64(time.Second))

		c.start = time.Now()
		c.end = c.start.Add(duration)

		ticker := time.NewTicker(config.timestep)
		timer := time.NewTimer(duration)

		c.updateColor(topic, durationSeconds)

		for {
			select {
			case <-ticker.C:
				c.updateColor(topic, durationSeconds)

			case <-timer.C:
				ticker.Stop()

				publish(c.client, topic, 1, false, ColorPayload(c.nextColor, 0))
				c.previousColor = c.nextColor

				continue outer
			}
		}
	}
}

func (c *ColorManager) updateColor(topic string, durationSeconds float64) {
	transition := min(max(time.Until(c.end), 0), config.timestep)

	elapsed := time.Since(c.start).Seconds()
	t := clamp01((elapsed + transition.Seconds()) / durationSeconds)

	colorStep := c.previousColor.Lerp(c.nextColor, t)
	publish(c.client, topic, 1, false, ColorPayload(colorStep, transition.Seconds()))
}

func chooseColor() HSB {
	choice := rand.IntN(2)
	switch choice {
	case 0:
		return config.a
	case 1:
		return config.b
	default:
		panic("invalid choice")
	}
}
