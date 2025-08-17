package main

import "math"

// HSB is Hue/Saturation/Brightness (a.k.a HSV).
// Hue is degrees [0..360), Saturation and Brightness are [0..1].
type HSB struct {
	H, S, B float64
}

// Lerp linearly interpolates between two HSB values.
// Hue wraps correctly around 360.
func (c HSB) Lerp(to HSB, t float64) HSB {
	// shortest hue delta
	dh := math.Mod(to.H-c.H+540, 360) - 180
	h := wrap360(c.H + dh*t)
	s := c.S + (to.S-c.S)*t
	b := c.B + (to.B-c.B)*t
	return HSB{H: h, S: s, B: b}
}

// ToSRGB converts HSB -> sRGB [0..1].
func (c HSB) ToSRGB() (r, g, b float64) {
	h := wrap360(c.H) / 60
	i := math.Floor(h)
	f := h - i
	p := c.B * (1 - c.S)
	q := c.B * (1 - c.S*f)
	t := c.B * (1 - c.S*(1-f))

	switch int(i) % 6 {
	case 0:
		return c.B, t, p
	case 1:
		return q, c.B, p
	case 2:
		return p, c.B, t
	case 3:
		return p, q, c.B
	case 4:
		return t, p, c.B
	default:
		return c.B, p, q
	}
}

// FromSRGB makes an HSB from sRGB [0..1].
func HSBFromSRGB(r, g, b float64) HSB {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	d := max - min

	var h float64
	if d == 0 {
		h = 0
	} else {
		switch max {
		case r:
			h = math.Mod((g-b)/d, 6)
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h *= 60
		if h < 0 {
			h += 360
		}
	}
	var s float64
	if max == 0 {
		s = 0
	} else {
		s = d / max
	}
	return HSB{H: h, S: s, B: max}
}

// Distance is a perceptual-ish Euclidean distance in HSB space.
// Hue wraps around, so 359° and 1° are close.
func (c HSB) Distance(o HSB) float64 {
	dh := math.Mod(o.H-c.H+540, 360) - 180
	ds := o.S - c.S
	db := o.B - c.B
	return math.Sqrt(dh*dh/360.0/360.0 + ds*ds + db*db)
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}

	if x > 1 {
		return 1
	}

	return x
}

func wrap360(h float64) float64 {
	h = math.Mod(h, 360.0)
	if h < 0 {
		h += 360.0
	}

	return h
}
