package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
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

func (m *Manager) Start(c mqtt.Client, friendlyName string, cfg RuntimeConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	m.addCancel(cancel)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic for light", "friendlyName", friendlyName, "err", err, "stack", debug.Stack())
			}
		}()

		cm := &ColorManager{
			client:       c,
			friendlyName: friendlyName,
			cfg:          cfg,
		}

		cm.Run(ctx)
	}()
}

type ColorManager struct {
	client mqtt.Client

	friendlyName string
	cfg          RuntimeConfig

	previousColor, nextColor Oklch
	start, end               time.Time
}

func (c *ColorManager) Run(ctx context.Context) {
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", c.friendlyName)
	c.previousColor = c.cfg.SelectColor()

outer:
	for {
		c.nextColor = c.cfg.SelectColor()

		duration := c.cfg.Transition()

		c.start = time.Now()
		c.end = c.start.Add(duration)

		ticker := time.NewTicker(c.cfg.timestep)
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
				case <-time.After(c.cfg.Hold()):
					break
				}

				continue outer
			}
		}
	}
}

func (c *ColorManager) updateColor(topic string, durationSeconds float64) {
	transition := min(max(time.Until(c.end), 0), c.cfg.timestep)

	elapsed := time.Since(c.start).Seconds()
	t := clamp01((elapsed + transition.Seconds()) / durationSeconds)

	colorStep := c.previousColor.Lerp(c.nextColor, t)
	publish(c.client, topic, 1, false, ColorPayload(colorStep, transition.Seconds()))
}
