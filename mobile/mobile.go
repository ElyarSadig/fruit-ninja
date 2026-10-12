// Package mobile is the entry point for Ebitengine's Android .aar binding.
// There is intentionally no call to ebiten.RunGame: Android owns the UI loop.
package mobile

import (
	"runtime"

	"example.com/fruit-fury/internal/game"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
	// Ebitengine's SetGame panics on desktop. Keep `go test ./...` usable.
	if runtime.GOOS != "android" && runtime.GOOS != "ios" {
		return
	}
	g, err := game.NewGame()
	if err != nil {
		panic(err)
	}
	mobile.SetGame(g)
}

// Dummy makes this package eligible for mobile binding generation.
func Dummy() {}

// SetDataDirectory selects Android's app-private files directory.
// Called once from the Java Activity before the first frame.
func SetDataDirectory(directory string) {
	game.SetScoreDirectory(directory)
}
