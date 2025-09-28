package main

import (
	"math/rand/v2"
	"time"
)

type TimedOverlay struct {
	fadeInStart  time.Time
	fadeInEnd    time.Time
	fadeOutStart time.Time
	fadeOutEnd   time.Time
}

type TimeOverlay TimedOverlay

func (t *TimeOverlay) Mix() float64 {
	now := time.Now().In(loc)
	n := minutesSinceMidnight(now)

	fiS := minutesSinceMidnight(asToday(now, t.fadeInStart))
	fiE := minutesSinceMidnight(asToday(now, t.fadeInEnd))
	foS := minutesSinceMidnight(asToday(now, t.fadeOutStart))
	foE := minutesSinceMidnight(asToday(now, t.fadeOutEnd))

	// piecewise on the circular day:
	switch {
	case inArc(fiS, fiE, n): // fading in
		return fracAlong(fiS, fiE, n)
	case inArc(fiE, foS, n): // fully on
		return 1.0
	case inArc(foS, foE, n): // fading out
		return 1.0 - fracAlong(foS, foE, n)
	default: // fully off
		return 0.0
	}
}

type DateOverlay TimedOverlay

func (d *DateOverlay) Mix() float64 {
	now := time.Now().In(loc)
	y := now.Year()
	total := yearLength(y)

	fiS := ydayFromMMDD(y, d.fadeInStart)
	fiE := ydayFromMMDD(y, d.fadeInEnd)
	foS := ydayFromMMDD(y, d.fadeOutStart)
	foE := ydayFromMMDD(y, d.fadeOutEnd)
	n := now.YearDay() - 1 // 0-based

	switch {
	case inArcDays(fiS, fiE, n, total): // fading in
		return fracAlongDays(fiS, fiE, n, total)
	case inArcDays(fiE, foS, n, total): // fully on
		return 1.0
	case inArcDays(foS, foE, n, total): // fading out
		return 1.0 - fracAlongDays(foS, foE, n, total)
	default: // off
		return 0.0
	}
}

type Overlay struct {
	colors Colors
	time   *TimeOverlay
	date   *DateOverlay
}

func (o *Overlay) Mix() float64 {
	var timeMix float64
	if o.time == nil {
		timeMix = 1
	} else {
		timeMix = o.time.Mix()
	}

	var dateMix float64
	if o.date == nil {
		dateMix = 1
	} else {
		dateMix = o.date.Mix()
	}

	return timeMix * dateMix
}

type RuntimeConfig struct {
	timestep time.Duration
	ambients Colors
	overlays []Overlay

	transitionMin time.Duration
	transitionMax time.Duration
	holdMin       time.Duration
	holdMax       time.Duration
}

func (r *RuntimeConfig) Transition() time.Duration {
	seconds := r.transitionMin.Seconds() + rand.Float64()*(r.transitionMax.Seconds()-r.transitionMin.Seconds())
	return time.Duration(seconds * float64(time.Second))
}

func (r *RuntimeConfig) Hold() time.Duration {
	seconds := r.holdMin.Seconds() + rand.Float64()*(r.holdMax.Seconds()-r.holdMin.Seconds())
	return time.Duration(seconds * float64(time.Second))
}

func (r *RuntimeConfig) SelectColor() Oklch {
	color := r.ambients.Select()
	for _, overlay := range r.overlays {
		color = color.Lerp(overlay.colors.Select(), overlay.Mix())
	}

	return color
}
