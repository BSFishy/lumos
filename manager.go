package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var manager = &Manager{}

type Manager struct {
	mu sync.Mutex

	cancels []context.CancelFunc
}

func (m *Manager) Lock() {
	m.mu.Lock()
}

func (m *Manager) Unlock() {
	m.mu.Unlock()
}

func (m *Manager) addCancel(cancel context.CancelFunc) {
	m.cancels = append(m.cancels, cancel)
}

func (m *Manager) CancelAll() {
	for _, cancel := range m.cancels {
		cancel()
	}

	m.cancels = []context.CancelFunc{}
}

func (m *Manager) Start(c mqtt.Client, friendlyName string, groupConfig GroupConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	m.addCancel(cancel)

	go func() {
		cm := &ColorManager{
			client:       c,
			friendlyName: friendlyName,
			groupConfig:  groupConfig,
		}

		cm.Run(ctx)
	}()
}

type ColorManager struct {
	client mqtt.Client

	friendlyName string
	groupConfig  GroupConfig

	previousColor, nextColor Oklch
	start, end               time.Time
}

func (c *ColorManager) Run(ctx context.Context) {
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
			case <-ctx.Done():
				return

			case <-ticker.C:
				c.updateColor(topic, duration.Seconds())

			case <-timer.C:
				ticker.Stop()

				publish(c.client, topic, 1, false, ColorPayload(c.nextColor, 0))
				c.previousColor = c.nextColor

				select {
				case <-ctx.Done():
					return
				case <-time.After(c.groupConfig.Hold.Select()):
					break
				}

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
