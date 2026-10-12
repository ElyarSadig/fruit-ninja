package game

import (
	"image/color"
	"math"
)

type Vec2 struct{ X, Y float64 }

func (v Vec2) Add(w Vec2) Vec2    { return Vec2{v.X + w.X, v.Y + w.Y} }
func (v Vec2) Sub(w Vec2) Vec2    { return Vec2{v.X - w.X, v.Y - w.Y} }
func (v Vec2) Mul(s float64) Vec2 { return Vec2{v.X * s, v.Y * s} }
func (v Vec2) Length() float64    { return math.Hypot(v.X, v.Y) }

func clamp(x, low, high float64) float64 {
	if x < low {
		return low
	}
	if x > high {
		return high
	}
	return x
}

func mix(a, b, t float64) float64 { return a + (b-a)*t }

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	t = clamp(t, 0, 1)
	return color.RGBA{
		R: uint8(mix(float64(a.R), float64(b.R), t)),
		G: uint8(mix(float64(a.G), float64(b.G), t)),
		B: uint8(mix(float64(a.B), float64(b.B), t)),
		A: uint8(mix(float64(a.A), float64(b.A), t)),
	}
}

func fade(c color.RGBA, a float64) color.RGBA {
	c.A = uint8(clamp(float64(c.A)*a, 0, 255))
	return c
}

// segmentHitsCircle makes quick mouse movements count, even if the mouse skips
// straight across a fruit between two frames. This is NOT point-only picking.
func segmentHitsCircle(from, to, center Vec2, radius float64) bool {
	d := to.Sub(from)
	d2 := d.X*d.X + d.Y*d.Y
	if d2 < 0.00001 {
		return center.Sub(from).Length() <= radius
	}
	v := center.Sub(from)
	t := clamp((v.X*d.X+v.Y*d.Y)/d2, 0, 1)
	closest := from.Add(d.Mul(t))
	dx, dy := closest.X-center.X, closest.Y-center.Y
	return dx*dx+dy*dy <= radius*radius
}

func pointInRect(p Vec2, x, y, w, h float64) bool {
	return p.X >= x && p.X <= x+w && p.Y >= y && p.Y <= y+h
}
