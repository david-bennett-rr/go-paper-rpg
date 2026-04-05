package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/editor"
)

func main() {
	ebiten.SetWindowSize(1400, 900)
	ebiten.SetWindowTitle("Paper RPG Map Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app, err := editor.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
