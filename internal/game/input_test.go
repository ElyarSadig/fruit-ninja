package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPrimaryTouch(t *testing.T) {
	a, b := ebiten.TouchID(7), ebiten.TouchID(8)
	tests := []struct {
		name   string
		ids    []ebiten.TouchID
		old    ebiten.TouchID
		had    bool
		want   ebiten.TouchID
		active bool
		fresh  bool
	}{
		{name: "nothing pressed", ids: nil, active: false},
		{name: "first finger", ids: []ebiten.TouchID{a}, want: a, active: true, fresh: true},
		{name: "stay on old finger", ids: []ebiten.TouchID{b, a}, old: a, had: true, want: a, active: true},
		{name: "switch after release", ids: []ebiten.TouchID{b}, old: a, had: true, want: b, active: true, fresh: true},
		{name: "release all", old: a, had: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, active, fresh := primaryTouch(tt.ids, tt.old, tt.had)
			if got != tt.want || active != tt.active || fresh != tt.fresh {
				t.Fatalf("primaryTouch() = (%d,%v,%v); want (%d,%v,%v)", got, active, fresh, tt.want, tt.active, tt.fresh)
			}
		})
	}
}
