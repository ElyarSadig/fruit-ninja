// Fruit Fury desktop launcher. For Android, see the mobile package.
package main

import (
	"log"

	"example.com/fruit-fury/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g, err := game.NewGame()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("FRUIT FURY | The Juiciest Slice")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(game.FramesPerSecond)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
