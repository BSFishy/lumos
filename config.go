package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BSFishy/lumos/util"
)

// TODO: read config from file
var config = Config{
	Groups: map[string]GroupConfig{
		"lumos_primary": {
			Ambient: []Color{
				"oklch(57.7% 0.245 27.325)",
				"oklch(64.6% 0.222 41.116)",
				"oklch(64.8% 0.2 131.684)",
				"oklch(54.6% 0.245 262.881)",
				"oklch(54.1% 0.281 293.009)",
				"oklch(66.7% 0.295 322.15)",
			},
			Time: []TimeConfig{
				{
					Colors: []Color{
						// Orange
						"oklch(75% 0.183 55.934)",
						"oklch(70.5% 0.213 47.604)",
						// Red
						"oklch(70.4% 0.191 22.216)",
						"oklch(63.7% 0.237 25.331)",
						// Yellow
						"oklch(82.8% 0.189 84.429)",
						"oklch(85.2% 0.199 91.936)",
						"oklch(79.5% 0.184 86.047)",
						// Amber
						"oklch(76.9% 0.188 70.08)",
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
					Colors: []Color{
						// Orange
						"oklch(55.3% 0.195 38.402)",
						"oklch(64.6% 0.222 41.116)",
						"oklch(70.5% 0.213 47.604)",
						// Amber
						"oklch(55.5% 0.163 48.998)",
						"oklch(66.6% 0.179 58.318)",
						"oklch(76.9% 0.188 70.08)",

						// Fuchsia
						"oklch(51.8% 0.253 323.949)",
						"oklch(66.7% 0.295 322.15)",
						// Purple
						"oklch(49.6% 0.265 301.924)",
						"oklch(62.7% 0.265 303.9)",
						// Violet
						"oklch(49.1% 0.27 292.581)",
						"oklch(60.6% 0.25 292.717)",
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

type Color string

func (co Color) Evaluate() Oklab {
	c := string(co)
	if strings.HasPrefix(c, "#") {
		switch len(c) {
		case 7:
			r := util.Must(strconv.ParseUint(c[1:3], 16, 8))
			g := util.Must(strconv.ParseUint(c[3:5], 16, 8))
			b := util.Must(strconv.ParseUint(c[5:7], 16, 8))

			return OklabFromSRGB(float64(r)/255, float64(g)/255, float64(b)/255)

		case 4:
			r := util.Must(strconv.ParseUint(string([]byte{c[1]}), 16, 8))
			g := util.Must(strconv.ParseUint(string([]byte{c[2]}), 16, 8))
			b := util.Must(strconv.ParseUint(string([]byte{c[3]}), 16, 8))

			return OklabFromSRGB(float64(r)/15, float64(g)/15, float64(b)/15)

		default:
			panic(fmt.Sprintf("invalid hex color: %s", c))
		}
	}

	if strings.HasPrefix(c, "oklab(") && strings.HasSuffix(c, ")") {
		params := strings.FieldsFunc(c[6:len(c)-1], func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })

		Lpercent := strings.TrimSpace(params[0])
		util.Assert(strings.HasSuffix(Lpercent, "%"), "invalid oklab format")

		L := util.Must(strconv.ParseFloat(Lpercent[0:len(Lpercent)-1], 64)) / 100
		A := util.Must(strconv.ParseFloat(strings.TrimSpace(params[1]), 64))
		B := util.Must(strconv.ParseFloat(strings.TrimSpace(params[2]), 64))

		return Oklab{
			L: L,
			A: A,
			B: B,
		}
	}

	if strings.HasPrefix(c, "oklch(") && strings.HasSuffix(c, ")") {
		params := strings.FieldsFunc(c[6:len(c)-1], func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })

		Lpercent := strings.TrimSpace(params[0])
		util.Assert(strings.HasSuffix(Lpercent, "%"), "invalid oklab format")

		L := util.Must(strconv.ParseFloat(Lpercent[0:len(Lpercent)-1], 64)) / 100
		C := util.Must(strconv.ParseFloat(strings.TrimSpace(params[1]), 64))
		H := util.Must(strconv.ParseFloat(strings.TrimSpace(params[2]), 64))

		return OklabFromOklch(L, C, H)
	}

	panic(fmt.Sprintf("invalid color format: %s", c))
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
	Colors  []Color `json:"colors"`

	colors []Oklab
}

func (t *TimeConfig) getColors() []Oklab {
	if t.colors != nil {
		return t.colors
	}

	colors := make([]Oklab, len(t.Colors))
	for i, color := range t.Colors {
		colors[i] = color.Evaluate()
	}

	t.colors = colors
	return colors
}

func (t *TimeConfig) SelectColor() (Oklab, float64) {
	// fallback if no colors configured
	if len(t.Colors) == 0 {
		return Oklab{}, 0
	}
	col := t.getColors()[rand.IntN(len(t.Colors))]

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
	Colors  []Color   `json:"colors"`

	colors []Oklab
}

func (s *SeasonalConfig) getColors() []Oklab {
	if s.colors != nil {
		return s.colors
	}

	colors := make([]Oklab, len(s.Colors))
	for i, color := range s.Colors {
		colors[i] = color.Evaluate()
	}

	s.colors = colors
	return colors
}

func (s *SeasonalConfig) SelectColor() (Oklab, float64) {
	if len(s.Colors) == 0 {
		return Oklab{}, 0
	}
	col := s.getColors()[rand.IntN(len(s.Colors))]

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
	Ambient    []Color          `json:"ambient"`
	Time       []TimeConfig     `json:"time"`
	Seasonal   []SeasonalConfig `json:"seasonal"`
	Transition Transition       `json:"transition"`
	Hold       Transition       `json:"hold"`

	// evaluated
	ambientColors []Oklab
}

func (g *GroupConfig) colors() []Oklab {
	if g.ambientColors != nil {
		return g.ambientColors
	}

	colors := make([]Oklab, len(g.Ambient))
	for i, color := range g.Ambient {
		colors[i] = color.Evaluate()
	}

	g.ambientColors = colors
	return colors
}

func (g *GroupConfig) SelectColor() Oklab {
	color := g.colors()[rand.IntN(len(g.Ambient))]

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
