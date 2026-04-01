package main

import (
	"fmt"
	"os"

	"rogue1980am/internal/game"
	"rogue1980am/internal/ui"
)

func main() {
	renderer, err := ui.NewRenderer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialise terminal: %v\n", err)
		os.Exit(1)
	}
	defer renderer.Close()

	engine := game.NewEngine(renderer)
	engine.RunMainMenu()
}
