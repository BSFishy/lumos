package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/BSFishy/lumos/util"
)

// TODO: read config from file
var config = Config{
	Groups: map[string]GroupConfig{
		"lumos_primary": {
			Ambient: []Oklab{
				OklabFromOklch(0.577, 0.245, 27.324),
				OklabFromOklch(0.768, 0.233, 120.85),
				OklabFromOklch(0.606, 0.25, 292.717),
				OklabFromOklch(0.795, 0.184, 86.047),
			},
			Transition: Transition{
				Minimum: "5s",
				Maximum: "10s",
			},
			Hold: Transition{
				Minimum: "0s",
				Maximum: "5s",
			},
		},
	},
	Timestep: "1s",
}

type Transition struct {
	Minimum string `json:"min"`
	Maximum string `json:"max"`
}

func (t *Transition) Min() time.Duration {
	return util.Must(time.ParseDuration(t.Minimum))
}

func (t *Transition) Max() time.Duration {
	return util.Must(time.ParseDuration(t.Maximum))
}

func (t *Transition) Select() time.Duration {
	seconds := t.Min().Seconds() + rand.Float64()*(t.Max().Seconds()-t.Min().Seconds())
	return time.Duration(seconds * float64(time.Second))
}

type GroupConfig struct {
	Priority   uint       `json:"priority"`
	Ambient    []Oklab    `json:"ambient"`
	Transition Transition `json:"transition"`
	Hold       Transition `json:"hold"`
}

func (g *GroupConfig) SelectColor() Oklab {
	color := g.Ambient[rand.IntN(len(g.Ambient))]

	// TODO: overlay time color and season color

	return color
}

type Config struct {
	Groups   map[string]GroupConfig `json:"groups"`
	Timestep string                 `json:"timestep"`
}

func (c *Config) TimestepDuration() time.Duration {
	return util.Must(time.ParseDuration(c.Timestep))
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
