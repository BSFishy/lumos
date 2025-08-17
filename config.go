package main

import (
	"encoding/json"
	"fmt"
	"time"
)

var config = Config{
	a: OklabFromSRGB(0.2, 0.5, 0.8),
	b: OklabFromSRGB(0.8, 0.7, 0.3),

	speedMin: 5 * time.Second,
	speedMax: 10 * time.Second,
	timestep: 1000 * time.Millisecond,
}

type Config struct {
	a, b     Oklab
	speedMin time.Duration
	speedMax time.Duration
	timestep time.Duration
}

func ColorPayload(color Oklab, transition float64) []byte {
	type Color struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	type Payload struct {
		Color      Color   `json:"color"`
		Transition float64 `json:"transition,omitempty"`
	}

	x, y := color.ToXY()

	payload := Payload{
		Color: Color{
			X: x,
			Y: y,
		},
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
