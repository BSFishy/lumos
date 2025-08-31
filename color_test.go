package main

import (
    "math"
    "testing"
)

func almostEqual(a, b float64) bool {
    const eps = 1e-5
    return math.Abs(a-b) < eps
}

func TestOklchLerp(t *testing.T) {
    from := Oklch{L: 0.2, C: 0.4, H: 350}
    to := Oklch{L: 0.8, C: 0.6, H: 10}

    got := from.Lerp(to, 0.5)
    if !almostEqual(got.L, 0.5) || !almostEqual(got.C, 0.5) || !almostEqual(got.H, 0) {
        t.Fatalf("lerp mid: got %#v", got)
    }

    if g := from.Lerp(to, 0); !almostEqual(g.L, from.L) || !almostEqual(g.C, from.C) || !almostEqual(g.H, from.H) {
        t.Fatalf("lerp t=0: got %#v", g)
    }

    if g := from.Lerp(to, 1); !almostEqual(g.L, to.L) || !almostEqual(g.C, to.C) || !almostEqual(g.H, to.H) {
        t.Fatalf("lerp t=1: got %#v", g)
    }
}

func TestOklchWaypoints(t *testing.T) {
    from := Oklch{L: 0, C: 0, H: 0}
    to := Oklch{L: 1, C: 1, H: 180}
    steps := 5
    pts := OklchWaypoints(from, to, steps)
    if len(pts) != steps {
        t.Fatalf("expected %d waypoints, got %d", steps, len(pts))
    }
    for i := 0; i < steps; i++ {
        expected := from.Lerp(to, float64(i)/float64(steps-1))
        got := pts[i]
        if !almostEqual(got.L, expected.L) || !almostEqual(got.C, expected.C) || !almostEqual(got.H, expected.H) {
            t.Fatalf("waypoint %d: got %#v want %#v", i, got, expected)
        }
    }

    small := OklchWaypoints(from, to, 1)
    if len(small) != 2 {
        t.Fatalf("expected 2 waypoints when steps<2, got %d", len(small))
    }
    if !almostEqual(small[0].L, from.L) || !almostEqual(small[1].L, to.L) {
        t.Fatalf("unexpected endpoints %#v", small)
    }
}

func TestSRGBRoundTrip(t *testing.T) {
    r, g, b := 0.25, 0.5, 0.75
    o := OklchFromSRGB(r, g, b)
    rr, gg, bb := o.ToSRGB()
    if !almostEqual(r, rr) || !almostEqual(g, gg) || !almostEqual(b, bb) {
        t.Fatalf("round trip mismatch: got %f %f %f", rr, gg, bb)
    }
}

func TestToHSV(t *testing.T) {
    o := OklchFromSRGB(1, 0, 0)
    h, s, v := o.ToHSV()
    if !almostEqual(h, 0) || !almostEqual(s, 1) || !almostEqual(v, 1) {
        t.Fatalf("HSV for red: got %f %f %f", h, s, v)
    }
}

func TestToXY(t *testing.T) {
    o := OklchFromSRGB(1, 1, 1)
    x, y := o.ToXY()
    if !almostEqual(x, 0.312727) || !almostEqual(y, 0.329023) {
        t.Fatalf("xy for white: got %f %f", x, y)
    }
}

