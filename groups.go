package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	primaryGroup []string             = nil
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
		var groups []Z2MGroup
		if err := json.Unmarshal(m.Payload(), &groups); err != nil {
			panic(fmt.Errorf("failed to unmarshal groups: %w", err))
		}

		for _, group := range groups {
			if group.FriendlyName == "lumos_primary" {
				addresses := []string{}
				for _, member := range group.Members {
					addresses = append(addresses, member.IeeeAddress)
				}

				primaryGroup = addresses
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
	if primaryGroup == nil || devices == nil {
		return
	}

	for _, address := range primaryGroup {
		friendlyName := devices[address].FriendlyName

		slog.Info("watching device", "friendly_name", friendlyName)
		go manager.Start(c, friendlyName)
	}
}
