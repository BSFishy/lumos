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
	steps    uint
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
	for _, overlay := range r.overlays {
		// a mix of 1 means this overlay should always take over. rand.Float64()
		// returns a value [0, 1) so that means if the mix is 1, a random value will
		// always be less than it. similarly, if the mix is 0, this overlay should
		// never be used. that means a random value will always be above the mix.
		// anything in between will randomly select this overlay or move to the next
		// one.
		threshold := overlay.Mix()
		r := rand.Float64()
		if r < threshold {
			return overlay.colors.Select()
		}
	}

	return r.ambients.Select()
}
