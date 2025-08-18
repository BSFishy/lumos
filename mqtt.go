package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/BSFishy/lumos/util"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func randSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func SetupMqtt() {
	broker, ok := os.LookupEnv("MQTT_BROKER")
	util.Assert(ok, "please specify an mqtt broker")

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)

	// Unique client ID, every run
	opts.SetClientID("lumos-" + randSuffix())

	// Robust connection behavior
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	opts.SetCleanSession(true) // we're resubscribing on connect

	// Reasonable keepalive
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetWriteTimeout(10 * time.Second)

	// (re)subscribe every time we reconnect
	opts.OnConnect = setupGroups

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
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
}
