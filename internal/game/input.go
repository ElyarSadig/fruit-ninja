package game

import (
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// primaryTouch sticks with the current finger even if other fingers join.
// A newly selected finger must start a new blade trail, not connect to the old one.
func primaryTouch(ids []ebiten.TouchID, old ebiten.TouchID, hadTouch bool) (id ebiten.TouchID, active, newFinger bool) {
	if hadTouch {
		for _, candidate := range ids {
			if candidate == old {
				return old, true, false
			}
		}
	}
	if len(ids) == 0 {
		return 0, false, false
	}
	return ids[0], true, true
}

// readPointer unifies touchscreen drags and desktop mouse swipes.
// Touch IDs are preserved across frames so multitouch can't cause phantom slashes.
func (g *Game) readPointer() (pos Vec2, pressed, down bool) {
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	id, active, fresh := primaryTouch(g.touchIDs, g.touchID, g.activeTouch)
	if active {
		if fresh {
			g.held = false
			g.trail = g.trail[:0]
		}
		x, y := ebiten.TouchPositionF(id)
		g.touchID = id
		g.activeTouch = true
		return Vec2{x, y}, fresh, true
	}
	if g.activeTouch {
		g.activeTouch = false
		// Ignore mouse emulation on the finger-release frame.
		return g.cursor, false, false
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		// Some Android devices synthesize a mouse click for a touch release.
		return g.cursor, false, false
	}
	x, y := ebiten.CursorPosition()
	return Vec2{float64(x), float64(y)},
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}
