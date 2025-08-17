package main

import (
	"encoding/json"
	"fmt"
	"time"
)

var config = Config{
	a: HSB{
		H: 279,
		S: 0.9,
		B: 1,
	},
	b: HSB{
		H: 89,
		S: 0.2,
		B: 1,
	},

	speedMin: 5 * time.Second,
	speedMax: 10 * time.Second,
	timestep: 1000 * time.Millisecond,
}

type Config struct {
	a, b     HSB
	speedMin time.Duration
	speedMax time.Duration
	timestep time.Duration
}

func ColorPayload(color HSB, transition float64) []byte {
	type Color struct {
		Hue        int `json:"hue"`
		Saturation int `json:"saturation"`
	}

	type Payload struct {
		Color      Color   `json:"color"`
		Brightness int     `json:"brightness"`
		Transition float64 `json:"transition,omitempty"`
	}

	payload := Payload{
		Color: Color{
			Hue:        int(color.H),
			Saturation: int(100 * color.S),
		},
		Brightness: int(256 * color.B),
	}

	if transition > 0 {
		payload.Transition = transition
	}

	data, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("failed to marshal color payload: %w", err))
	}

	return data
}
