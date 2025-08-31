package main

import "math"

type Oklch struct{ L, C, H float64 } // L in [0..1], H in degrees [0..360)

// ---------- Blending ----------
func (c Oklch) Lerp(to Oklch, t float64) Oklch {
	return Oklch{
		L: c.L + (to.L-c.L)*t,
		C: c.C + (to.C-c.C)*t,
		H: lerpHue(c.H, to.H, t),
	}
}

// Generate N waypoints along the Oklch line (inclusive of endpoints).
// Use these as keyframes; send each with a per-step `transition`.
func OklchWaypoints(from, to Oklch, steps int) []Oklch {
	if steps < 2 {
		steps = 2
	}
	out := make([]Oklch, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		out[i] = from.Lerp(to, t)
	}
	return out
}

func lerpHue(a, b, t float64) float64 {
	d := math.Mod(b-a+180, 360) - 180
	h := a + d*t
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}

// ---------- sRGB <-> Oklch ----------
func OklchFromSRGB(r, g, b float64) Oklch {
	rl, gl, bl := srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)

	// linear RGB -> LMS
	l := 0.4122214708*rl + 0.5363325363*gl + 0.0514459929*bl
	m := 0.2119034982*rl + 0.6806995451*gl + 0.1073969566*bl
	s := 0.0883024619*rl + 0.2817188376*gl + 0.6299787005*bl

	l_ := cbrt(l)
	m_ := cbrt(m)
	s_ := cbrt(s)

	L := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	A := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	B := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

	h := math.Atan2(B, A) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return Oklch{
		L: L,
		C: math.Hypot(A, B),
		H: h,
	}
}

func (c Oklch) ToSRGB() (r, g, b float64) {
	h := c.H * math.Pi / 180.0
	a := c.C * math.Cos(h)
	b_ := c.C * math.Sin(h)

	// Oklch -> LMS'
	l_ := c.L + 0.3963377774*a + 0.2158037573*b_
	m_ := c.L - 0.1055613458*a - 0.0638541728*b_
	s_ := c.L - 0.0894841775*a - 1.2914855480*b_

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	// LMS -> linear RGB
	rl := +4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	gl := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bl := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	// linear -> sRGB with clamp
	r = clamp01(linearToSrgb(rl))
	g = clamp01(linearToSrgb(gl))
	b = clamp01(linearToSrgb(bl))
	return
}

// ---------- Helpers (HSV + xy if you need them) ----------

// ToHSB/HSV via sRGB (h in [0,360), s,v in [0,1])
func (c Oklch) ToHSV() (h, s, v float64) {
	r, g, b := c.ToSRGB()
	return rgbToHSV(r, g, b)
}

// CIE 1931 xy (sRGB/D65). If all zero, returns 0,0.
func (c Oklch) ToXY() (x, y float64) {
	r, g, b := c.ToSRGB()
	rl, gl, bl := srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)
	X := 0.4124564*rl + 0.3575761*gl + 0.1804375*bl
	Y := 0.2126729*rl + 0.7151522*gl + 0.0721750*bl
	Z := 0.0193339*rl + 0.1191920*gl + 0.9503041*bl
	sum := X + Y + Z
	if sum == 0 {
		return 0, 0
	}
	return X / sum, Y / sum
}

// ---------- Private utilities ----------
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func linearToSrgb(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func cbrt(x float64) float64 {
	if x > 0 {
		return math.Pow(x, 1.0/3.0)
	}
	if x < 0 {
		return -math.Pow(-x, 1.0/3.0)
	}
	return 0
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

// Minimal RGB->HSV (for ToHSV)
func rgbToHSV(r, g, b float64) (h, s, v float64) {
	max := r
	if g > max {
		max = g
	}
	if b > max {
		max = b
	}
	min := r
	if g < min {
		min = g
	}
	if b < min {
		min = b
	}
	v = max
	d := max - min
	if max == 0 {
		return 0, 0, 0
	}
	s = d / max
	if d == 0 {
		return 0, s, v
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h >= 360 {
		h -= 360
	}
	return
}
