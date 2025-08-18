package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
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
			Time: []TimeConfig{
				{
					Colors: []Oklab{
						OklabFromOklch(0.901, 0.076, 70.697),
						OklabFromOklch(0.75, 0.183, 55.934),
						OklabFromOklch(0.885, 0.062, 18.334),
						OklabFromOklch(0.704, 0.191, 22.216),
						OklabFromOklch(0.924, 0.12, 95.746),
						OklabFromOklch(0.828, 0.189, 84.429),
						OklabFromOklch(0.945, 0.129, 101.54),
					},
					FadeIn: Fader{
						Start: "8:00PM",
						End:   "10:00PM",
					},
					FadeOut: Fader{
						Start: "8:00AM",
						End:   "10:00AM",
					},
				},
			},
			Seasonal: []SeasonalConfig{
				// Halloween
				{
					Colors: []Oklab{
						// Orange
						OklabFromOklch(0.553, 0.195, 38.402),
						OklabFromOklch(0.75, 0.183, 55.934),
						OklabFromOklch(0.555, 0.163, 48.998),
						OklabFromOklch(0.666, 0.179, 58.318),
						OklabFromOklch(0.769, 0.188, 70.08),

						// Purple
						OklabFromOklch(0.518, 0.253, 323.949),
						OklabFromOklch(0.667, 0.295, 322.15),
						OklabFromOklch(0.496, 0.265, 301.924),
						OklabFromOklch(0.627, 0.265, 303.9),
						OklabFromOklch(0.491, 0.27, 292.581),
						OklabFromOklch(0.606, 0.25, 292.717),
					},
					FadeIn: DateFader{
						Start: "10-01",
						End:   "10-20",
					},
					FadeOut: DateFader{
						Start: "11-01",
						End:   "11-07",
					},
				},
			},
			Transition: Transition{
				Minimum: "5s",
				Maximum: "30s",
			},
			Hold: Transition{
				Minimum: "0s",
				Maximum: "15s",
			},
		},
	},
	Timestep: "1s",
}

var loc = mustLoadLocation()

func mustLoadLocation() *time.Location {
	if tz := os.Getenv("TZ"); tz != "" {
		l, err := time.LoadLocation(tz)
		if err == nil {
			return l
		}
	}
	// fallback: your local or a hardcoded zone you want
	l, _ := time.LoadLocation("America/Chicago")
	if l == nil {
		l = time.Local
	}
	return l
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

type Fader struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func (f *Fader) StartTime() time.Time {
	return util.Must(time.ParseInLocation(time.Kitchen, f.Start, loc))
}

func (f *Fader) EndTime() time.Time {
	return util.Must(time.ParseInLocation(time.Kitchen, f.End, loc))
}

type TimeConfig struct {
	FadeIn  Fader   `json:"fade_in"`
	FadeOut Fader   `json:"fade_out"`
	Colors  []Oklab `json:"colors"`
}

func (t *TimeConfig) SelectColor() (Oklab, float64) {
	// fallback if no colors configured
	if len(t.Colors) == 0 {
		return Oklab{}, 0
	}
	col := t.Colors[rand.IntN(len(t.Colors))]

	now := time.Now().In(loc)
	n := minutesSinceMidnight(now)

	fiS := minutesSinceMidnight(asToday(now, t.FadeIn.StartTime()))
	fiE := minutesSinceMidnight(asToday(now, t.FadeIn.EndTime()))
	foS := minutesSinceMidnight(asToday(now, t.FadeOut.StartTime()))
	foE := minutesSinceMidnight(asToday(now, t.FadeOut.EndTime()))

	// piecewise on the circular day:
	switch {
	case inArc(fiS, fiE, n): // fading in
		return col, fracAlong(fiS, fiE, n)
	case inArc(fiE, foS, n): // fully on
		return col, 1.0
	case inArc(foS, foE, n): // fading out
		return col, 1.0 - fracAlong(foS, foE, n)
	default: // fully off
		return col, 0.0
	}
}

type DateFader struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

var dayLayout = "01-02"

func (d *DateFader) StartDate() time.Time {
	return util.Must(time.ParseInLocation(dayLayout, d.Start, loc))
}

func (d *DateFader) EndDate() time.Time {
	return util.Must(time.ParseInLocation(dayLayout, d.End, loc))
}

type SeasonalConfig struct {
	FadeIn  DateFader `json:"fade_in"`
	FadeOut DateFader `json:"fade_out"`
	Colors  []Oklab   `json:"colors"`
}

func (s *SeasonalConfig) SelectColor() (Oklab, float64) {
	if len(s.Colors) == 0 {
		return Oklab{}, 0
	}
	col := s.Colors[rand.IntN(len(s.Colors))]

	now := time.Now().In(loc)
	y := now.Year()
	total := yearLength(y)

	fiS := ydayFromMMDD(y, s.FadeIn.StartDate())
	fiE := ydayFromMMDD(y, s.FadeIn.EndDate())
	foS := ydayFromMMDD(y, s.FadeOut.StartDate())
	foE := ydayFromMMDD(y, s.FadeOut.EndDate())
	n := now.YearDay() - 1 // 0-based

	switch {
	case inArcDays(fiS, fiE, n, total): // fading in
		return col, fracAlongDays(fiS, fiE, n, total)
	case inArcDays(fiE, foS, n, total): // fully on
		return col, 1.0
	case inArcDays(foS, foE, n, total): // fading out
		return col, 1.0 - fracAlongDays(foS, foE, n, total)
	default: // off
		return col, 0.0
	}
}

type GroupConfig struct {
	Priority   uint             `json:"priority"`
	Ambient    []Oklab          `json:"ambient"`
	Time       []TimeConfig     `json:"time"`
	Seasonal   []SeasonalConfig `json:"seasonal"`
	Transition Transition       `json:"transition"`
	Hold       Transition       `json:"hold"`
}

func (g *GroupConfig) SelectColor() Oklab {
	color := g.Ambient[rand.IntN(len(g.Ambient))]

	for _, seasonConfig := range g.Seasonal {
		color = color.Lerp(seasonConfig.SelectColor())
	}

	for _, timeConfig := range g.Time {
		color = color.Lerp(timeConfig.SelectColor())
	}

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

// minutes since midnight using local clock
func minutesSinceMidnight(t time.Time) int {
	t = t.Local()
	return t.Hour()*60 + t.Minute()
}

// take clock time (hour/minute) from src, anchor it to the date of base (today)
func asToday(base, src time.Time) time.Time {
	y, m, d := base.Date()
	loc := base.Location()
	return time.Date(y, m, d, src.Hour(), src.Minute(), 0, 0, loc)
}

// is x on the circular arc going forward from a to b (inclusive) on [0,1440)
func inArc(a, b, x int) bool {
	return distFwd(a, x) <= distFwd(a, b)
}

// forward modular distance from a → b in minutes (0..1439)
func distFwd(a, b int) int {
	const day = 24 * 60
	d := (b - a) % day
	if d < 0 {
		d += day
	}
	return d
}

// fractional position of x on arc a→b (handles wrap). If a==b, returns 1.
func fracAlong(a, b, x int) float64 {
	total := distFwd(a, b)
	if total == 0 {
		return 1.0
	}
	return float64(distFwd(a, x)) / float64(total)
}

func yearLength(y int) int {
	if (y%4 == 0 && y%100 != 0) || (y%400 == 0) {
		return 366
	}
	return 365
}

// Parse "MM-DD" and convert to 0-based yearday for given year.
func ydayFromMMDD(year int, t time.Time) int {
	tt := time.Date(year, t.Month(), t.Day(), 0, 0, 0, 0, loc)
	return tt.YearDay() - 1 // 0-based
}

// Is x on the forward arc a→b (inclusive) on a 0..total-1 ring.
func inArcDays(a, b, x, total int) bool {
	return distFwdDays(a, x, total) <= distFwdDays(a, b, total)
}

// Forward modular distance a→b.
func distFwdDays(a, b, total int) int {
	d := (b - a) % total
	if d < 0 {
		d += total
	}
	return d
}

// Fractional position of x on arc a→b (handles wrap). If a==b, treat as instant (1).
func fracAlongDays(a, b, x, total int) float64 {
	tot := distFwdDays(a, b, total)
	if tot == 0 {
		return 1.0
	}
	return float64(distFwdDays(a, x, total)) / float64(tot)
}
