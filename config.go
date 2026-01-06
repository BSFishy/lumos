package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BSFishy/lumos/util"
)

func SetupConfig() {
	contents, err := os.ReadFile("/config/config.json")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("failed to read config file: %w", err))
		}

		slog.Info("no config file found")
		return
	}

	var data Config
	if err = json.Unmarshal(contents, &data); err != nil {
		slog.Warn("failed to decode config, using default", "err", err)
		return
	}

	config = data
}

var config = Config{
	Groups: []GroupConfig{},
	Steps:  5,
}

type Color string

func (co Color) Evaluate() Oklch {
	c := string(co)
	if strings.HasPrefix(c, "#") {
		switch len(c) {
		case 7:
			r := util.Must(strconv.ParseUint(c[1:3], 16, 8))
			g := util.Must(strconv.ParseUint(c[3:5], 16, 8))
			b := util.Must(strconv.ParseUint(c[5:7], 16, 8))

			return OklchFromSRGB(float64(r)/255, float64(g)/255, float64(b)/255)

		case 4:
			r := util.Must(strconv.ParseUint(string([]byte{c[1]}), 16, 8))
			g := util.Must(strconv.ParseUint(string([]byte{c[2]}), 16, 8))
			b := util.Must(strconv.ParseUint(string([]byte{c[3]}), 16, 8))

			return OklchFromSRGB(float64(r)/15, float64(g)/15, float64(b)/15)

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

		H := math.Atan2(B, A) * 180 / math.Pi
		if H < 0 {
			H += 360
		}
		return Oklch{L: L, C: math.Hypot(A, B), H: H}
	}

	if strings.HasPrefix(c, "oklch(") && strings.HasSuffix(c, ")") {
		params := strings.FieldsFunc(c[6:len(c)-1], func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })

		Lpercent := strings.TrimSpace(params[0])
		util.Assert(strings.HasSuffix(Lpercent, "%"), "invalid oklab format")

		L := util.Must(strconv.ParseFloat(Lpercent[0:len(Lpercent)-1], 64)) / 100
		C := util.Must(strconv.ParseFloat(strings.TrimSpace(params[1]), 64))
		H := util.Must(strconv.ParseFloat(strings.TrimSpace(params[2]), 64))

		return Oklch{L: L, C: C, H: H}
	}

	panic(fmt.Sprintf("invalid color format: %s", c))
}

type Colors struct {
	colors             []Oklch
	previouslySelected uint
}

func (c *Colors) Select() Oklch {
	util.Assert(len(c.colors) > 0, "must have colors")

	idx := rand.UintN(uint(len(c.colors)) - 1)
	if idx == c.previouslySelected {
		idx = uint(len(c.colors)) - 1
	}

	col := c.colors[idx]
	c.previouslySelected = idx
	return col
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

type TimeFader struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type TimeConfig struct {
	FadeIn  TimeFader `json:"fade_in"`
	FadeOut TimeFader `json:"fade_out"`
}

func (t TimeConfig) Compile() *TimeOverlay {
	return &TimeOverlay{
		fadeInStart:  util.Must(time.ParseInLocation(time.Kitchen, t.FadeIn.Start, loc)),
		fadeInEnd:    util.Must(time.ParseInLocation(time.Kitchen, t.FadeIn.End, loc)),
		fadeOutStart: util.Must(time.ParseInLocation(time.Kitchen, t.FadeOut.Start, loc)),
		fadeOutEnd:   util.Must(time.ParseInLocation(time.Kitchen, t.FadeOut.End, loc)),
	}
}

type DateFader struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type SeasonalConfig struct {
	FadeIn  DateFader `json:"fade_in"`
	FadeOut DateFader `json:"fade_out"`
}

var dayLayout = "01-02"

func (s SeasonalConfig) Compile() *DateOverlay {
	return &DateOverlay{
		fadeInStart:  util.Must(time.ParseInLocation(dayLayout, s.FadeIn.Start, loc)),
		fadeInEnd:    util.Must(time.ParseInLocation(dayLayout, s.FadeIn.End, loc)),
		fadeOutStart: util.Must(time.ParseInLocation(dayLayout, s.FadeOut.Start, loc)),
		fadeOutEnd:   util.Must(time.ParseInLocation(dayLayout, s.FadeOut.End, loc)),
	}
}

type GroupConfig struct {
	Colors    []Color  `json:"colors"`
	AppliesTo []string `json:"applies_to"`

	Time *TimeConfig     `json:"time"`
	Date *SeasonalConfig `json:"date"`
}

func (g *GroupConfig) Contains(groups []string) bool {
	if len(g.AppliesTo) == 0 {
		return true
	}

	for _, group := range g.AppliesTo {
		if slices.Contains(groups, group) {
			return true
		}
	}

	return false
}

func (g *GroupConfig) IsAmbient() bool {
	return g.Time == nil && g.Date == nil
}

func (g *GroupConfig) CompileColors() []Oklch {
	colors := make([]Oklch, len(g.Colors))
	for i, color := range g.Colors {
		colors[i] = color.Evaluate()
	}

	return colors
}

type Transition struct {
	Minimum string `json:"min"`
	Maximum string `json:"max"`
}

type Config struct {
	Steps      uint       `json:"steps"`
	Transition Transition `json:"transition"`
	Hold       Transition `json:"hold"`

	Groups []GroupConfig `json:"groups"`
}

func (c *Config) ContainsGroup(name string) bool {
	list := []string{name}
	for _, group := range c.Groups {
		if group.Contains(list) {
			return true
		}
	}

	return false
}

func (c *Config) Compile(groups []string) RuntimeConfig {
	ambients := []Oklch{}
	overlays := []Overlay{}

	for _, group := range c.Groups {
		if !group.Contains(groups) {
			continue
		}

		if group.IsAmbient() {
			ambients = append(ambients, group.CompileColors()...)
			continue
		}

		var timeOverlay *TimeOverlay
		if group.Time != nil {
			timeOverlay = group.Time.Compile()
		}

		var dateOverlay *DateOverlay
		if group.Date != nil {
			dateOverlay = group.Date.Compile()
		}

		overlays = append(overlays, Overlay{
			time: timeOverlay,
			date: dateOverlay,
			colors: Colors{
				colors: group.CompileColors(),
			},
		})
	}

	return RuntimeConfig{
		steps: c.Steps,
		ambients: Colors{
			colors: ambients,
		},
		overlays: overlays,

		transitionMin: util.Must(time.ParseDuration(c.Transition.Minimum)),
		transitionMax: util.Must(time.ParseDuration(c.Transition.Maximum)),
		holdMin:       util.Must(time.ParseDuration(c.Hold.Minimum)),
		holdMax:       util.Must(time.ParseDuration(c.Hold.Maximum)),
	}
}

func ColorPayload(color Oklch, transition float64) []byte {
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
