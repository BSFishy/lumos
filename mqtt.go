package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/BSFishy/lumos/util"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func SetupMqtt() mqtt.Client {
	broker, ok := os.LookupEnv("MQTT_BROKER")
	util.Assert(ok, "please specify an mqtt broker")

	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("lumos")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return c
}

func subscribe(c mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) {
	token := c.Subscribe(topic, qos, callback)

	if token.Wait() && token.Error() != nil {
		panic(fmt.Errorf("failed to subscribe to %s: %w", topic, token.Error()))
	}
}

func publish(c mqtt.Client, topic string, qos byte, retained bool, payload any) {
	token := c.Publish(topic, qos, retained, payload)

	if token.Wait() && token.Error() != nil {
		panic(fmt.Errorf("failed to publish to %s: %w", topic, token.Error()))
	}

	slog.Debug("publish", "topic", topic, "payload", string(payload.([]byte)))
}
