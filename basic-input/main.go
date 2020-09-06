package main

import (
	"errors"
	"fmt"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	translateFactor = 10
	screenWidth     = 640
	screenHeight    = 480
)

var (
	ErrCleanExit = errors.New("clean exit, no error")
)

// Sprite is from the ebiten drag and drop (drag) example.
type Sprite struct {
	id  string
	img *ebiten.Image
	x   int
	y   int
}

func (s *Sprite) In(x, y int) bool {
	// Check the actual color (alpha) value at the specified position
	// so that the result of In becomes natural to users.
	//
	// Note that this is not a good manner to use At for logic
	// since color from At might include some errors on some machines.
	// As this is not so important logic, it's ok to use it so far.
	return s.img.At(x-s.x, y-s.y).(color.RGBA).A > 0
}

// MoveBy moves the sprite by (x, y).
func (s *Sprite) MoveBy(x, y int) {
	w, h := s.img.Size()

	s.x += x
	s.y += y

	if s.x < 0 {
		s.x = 0
	}

	if s.x > screenWidth-w {
		s.x = screenWidth - w
	}

	if s.y < 0 {
		s.y = 0
	}

	if s.y > screenHeight-h {
		s.y = screenHeight - h
	}
}

func (s *Sprite) Draw(screen *ebiten.Image, dx, dy int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+dx), float64(s.y+dy))
	screen.DrawImage(s.img, op)
}

type Game struct {
	s            []*Sprite
	activeSprite int
}

func (g *Game) Update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.s[g.activeSprite].MoveBy(0, -translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.s[g.activeSprite].MoveBy(0, translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.s[g.activeSprite].MoveBy(-translateFactor, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.s[g.activeSprite].MoveBy(translateFactor, 0)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Because we draw in slice order, the latest is the one on top,
		// so check from latest to first
		for i := len(g.s) - 1; i >= 0; i-- {
			s := g.s[i]
			if s.In(cx, cy) {
				g.activeSprite = i

				break
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrCleanExit
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Active sprite: "+g.s[g.activeSprite].id)

	for _, s := range g.s {
		s.Draw(screen, 0, 0)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func main() {
	img, _, err := ebitenutil.NewImageFromFile("../images/gopher.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	g := &Game{
		s: []*Sprite{{"0", img, 0, 0}, {"1", img, 100, 100}},
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Basic Input")

	if err := ebiten.RunGame(g); err != nil {
		if errors.Is(err, ErrCleanExit) {
			fmt.Println("Good bye!")

			return
		}

		log.Fatal(err)
	}
}
