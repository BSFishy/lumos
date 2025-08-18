package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	deviceGroups map[string][]string  = nil
	devices      map[string]Z2MDevice = nil
)

type Z2MGroup struct {
	FriendlyName string           `json:"friendly_name"`
	ID           int              `json:"id"`
	Members      []Z2MGroupMember `json:"members"`
}

type Z2MGroupMember struct {
	Endpoint    int    `json:"endpoint"`
	IeeeAddress string `json:"ieee_address"`
}

type Z2MDevice struct {
	FriendlyName string `json:"friendly_name"`
	IeeeAddress  string `json:"ieee_address"`
}

func setupGroups(client mqtt.Client) {
	// get boot-time config
	subscribe(client, "zigbee2mqtt/bridge/groups", 1, onGroups)
	subscribe(client, "zigbee2mqtt/bridge/devices", 1, onDevices)

	// subscribe to replies
	subscribe(client, "zigbee2mqtt/bridge/config/groups", 1, onGroups)
	subscribe(client, "zigbee2mqtt/bridge/config/devices", 1, onDevices)

	// optional: watch for events (joins/leaves) then re-fetch
	subscribe(client, "zigbee2mqtt/bridge/event", 1, func(c mqtt.Client, m mqtt.Message) {
		// on device_joined/device_leave/groups_changed -> request fresh lists
		publish(c, "zigbee2mqtt/bridge/config/devices/get", 0, false, "")
		publish(c, "zigbee2mqtt/bridge/config/groups/get", 0, false, "")
	})

	// initial fetch
	publish(client, "zigbee2mqtt/bridge/config/groups/get", 0, false, "")
	publish(client, "zigbee2mqtt/bridge/config/devices/get", 0, false, "")
}

func onGroups(c mqtt.Client, m mqtt.Message) {
	var payloadGroups []Z2MGroup
	if err := json.Unmarshal(m.Payload(), &payloadGroups); err != nil {
		panic(fmt.Errorf("failed to unmarshal groups: %w", err))
	}

	deviceGroups = map[string][]string{}
	for _, group := range payloadGroups {
		if _, ok := config.Groups[group.FriendlyName]; ok {
			for _, member := range group.Members {
				deviceGroups[member.IeeeAddress] = append(deviceGroups[member.IeeeAddress], group.FriendlyName)
			}
		}
	}

	refreshDevices(c)
}

func onDevices(c mqtt.Client, m mqtt.Message) {
	var payloadDevices []Z2MDevice
	if err := json.Unmarshal(m.Payload(), &payloadDevices); err != nil {
		panic(fmt.Errorf("failed to unmarshal devices: %w", err))
	}

	devices = map[string]Z2MDevice{}
	for _, device := range payloadDevices {
		devices[device.IeeeAddress] = device
	}

	refreshDevices(c)
}

func refreshDevices(c mqtt.Client) {
	if deviceGroups == nil || devices == nil {
		return
	}

	manager.Lock()
	defer manager.Unlock()

	manager.CancelAll()

	for _, device := range devices {
		groups, ok := deviceGroups[device.IeeeAddress]
		if !ok {
			continue
		}

		groupName := groups[0]
		highestPriority := config.Groups[groupName].Priority
		for _, group := range groups[1:] {
			if priority := config.Groups[group].Priority; priority > highestPriority {
				groupName = group
				highestPriority = priority
			}
		}

		group := config.Groups[groupName]
		manager.Start(c, device.FriendlyName, group)

		slog.Info("controlling device", "friendly_name", device.FriendlyName, "group", groupName)
	}
}
