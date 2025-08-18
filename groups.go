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
	subscribe(client, "zigbee2mqtt/bridge/groups", 1, func(c mqtt.Client, m mqtt.Message) {
		var payloadGroups []Z2MGroup
		if err := json.Unmarshal(m.Payload(), &payloadGroups); err != nil {
			panic(fmt.Errorf("failed to unmarshal groups: %w", err))
		}

		deviceGroups = map[string][]string{}
		for _, group := range payloadGroups {
			if _, ok := config.Groups[group.FriendlyName]; ok {
				for _, member := range group.Members {
					groups, ok := deviceGroups[member.IeeeAddress]
					if !ok {
						groups = []string{}
					}

					groups = append(groups, group.FriendlyName)
					deviceGroups[member.IeeeAddress] = groups
				}
			}
		}

		refreshDevices(c)
	})

	subscribe(client, "zigbee2mqtt/bridge/devices", 1, func(c mqtt.Client, m mqtt.Message) {
		var payloadDevices []Z2MDevice
		if err := json.Unmarshal(m.Payload(), &payloadDevices); err != nil {
			panic(fmt.Errorf("failed to unmarshal devices: %w", err))
		}

		devices = map[string]Z2MDevice{}
		for _, device := range payloadDevices {
			devices[device.IeeeAddress] = device
		}

		refreshDevices(c)
	})
}

func refreshDevices(c mqtt.Client) {
	if deviceGroups == nil || devices == nil {
		return
	}

	// TODO: clear all running goroutines

	for _, device := range devices {
		groups, ok := deviceGroups[device.IeeeAddress]
		if !ok {
			continue
		}

		groupName := groups[0]
		highestPriorityGroup := config.Groups[groups[0]]
		highestPriority := highestPriorityGroup.Priority
		for _, group := range groups[1:] {
			configGroup := config.Groups[group]
			priority := configGroup.Priority

			if priority > highestPriority {
				groupName = group
				highestPriorityGroup = configGroup
				highestPriority = priority
			}
		}

		group := highestPriorityGroup
		go manager.Start(c, device.FriendlyName, group)

		slog.Info("controlling device", "friendly_name", device.FriendlyName, "group", groupName)
	}
}
