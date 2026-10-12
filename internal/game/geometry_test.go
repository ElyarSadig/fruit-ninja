package game

import "testing"

func TestSegmentHitsCircle(t *testing.T) {
	for _, tc := range []struct {
		name         string
		a, b, center Vec2
		radius       float64
		hit          bool
	}{
		{"straight-through", Vec2{0, 0}, Vec2{100, 0}, Vec2{50, 0}, 10, true},
		{"fast-diagonal", Vec2{-100, -100}, Vec2{100, 100}, Vec2{0, 0}, 18, true},
		{"far-away", Vec2{0, 0}, Vec2{100, 0}, Vec2{50, 30}, 10, false},
		{"touch-endpoint", Vec2{0, 0}, Vec2{10, 0}, Vec2{12, 0}, 2, true},
		{"zero-distance-inside", Vec2{3, 4}, Vec2{3, 4}, Vec2{3, 4}, 1, true},
		{"zero-distance-outside", Vec2{3, 4}, Vec2{3, 4}, Vec2{15, 20}, 1, false},
		{"not-infinite-ray", Vec2{0, 0}, Vec2{5, 0}, Vec2{40, 0}, 10, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := segmentHitsCircle(tc.a, tc.b, tc.center, tc.radius)
			if actual != tc.hit {
				t.Fatalf("got %v; want %v", actual, tc.hit)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	if clamp(-2, 0, 1) != 0 || clamp(4, 0, 1) != 1 || clamp(0.4, 0, 1) != 0.4 {
		t.Fatal("clamp bounds regression")
	}
}
