package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/editor"
)

func main() {
	ebiten.SetWindowSize(editor.DefaultWindowW, editor.DefaultWindowH)
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
