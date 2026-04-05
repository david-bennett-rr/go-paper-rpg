package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/app"
)

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("Paper RPG")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := app.NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
