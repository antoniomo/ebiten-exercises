package main

import (
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type Game struct {
	turn int
}

func (g *Game) Update(screen *ebiten.Image) error {
	// As a turn-based strategy, just register the player's declared
	// "actions" first, then trigger world update only if the "next turn"
	// trigger applies, otherwise skip
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.turn++
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Turn: "+strconv.Itoa(g.turn))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetMaxTPS(20) // Turn based so this is to reduce the load

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
