package main

import (
	"errors"
	"fmt"
	"image/color"
	_ "image/png"
	"log"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	translateFactor = 10
	rotateFactor    = 0.05
	screenWidth     = 640
	screenHeight    = 480
)

var (
	ErrCleanExit = errors.New("clean exit, no error")
)

func genCircle(r int, clr color.Color) *ebiten.Image {
	dc := gg.NewContext(r*2, r*2)
	dc.DrawCircle(float64(r), float64(r), float64(r))
	dc.SetColor(clr)
	dc.Fill()

	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	return img
}

func genRectangle(w, h int, clr color.Color) *ebiten.Image {
	dc := gg.NewContext(w, h)
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.SetColor(clr)
	dc.Fill()

	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	return img
}

func genPolygon(n, r int, clr color.Color) *ebiten.Image {
	dc := gg.NewContext(r*2, r*2)
	dc.DrawRegularPolygon(n, float64(r), float64(r), float64(r), 0)
	dc.SetColor(clr)
	dc.Fill()

	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	return img
}

type Shape struct {
	id    string
	x     int
	y     int
	theta float64
	img   *ebiten.Image
}

func NewShape(id string, x, y int, theta float64, img *ebiten.Image) *Shape {
	s := &Shape{
		id:    id,
		x:     x,
		y:     y,
		theta: theta,
		img:   img,
	}

	return s
}

// In is from the ebiten drag and drop (drag) example.
func (s *Shape) In(x, y int) bool {
	w, h := s.img.Size()

	return s.img.At(x-s.x+w, y-s.y+h).(color.RGBA).A > 0
}

// MoveBy moves the shape by (x, y).
func (s *Shape) MoveBy(x, y int) {
	s.x += x
	s.y += y
	w, h := s.img.Size()

	if s.x < 0+w {
		s.x = 0 + w
	}

	if s.x > screenWidth-w {
		s.x = screenWidth - w
	}

	if s.y < 0+h {
		s.y = 0 + h
	}

	if s.y > screenHeight-h {
		s.y = screenHeight - h
	}
}

func (s *Shape) Draw(screen *ebiten.Image) {
	w, h := s.img.Size()

	op := &ebiten.DrawImageOptions{}
	// From Ebiten's rotate example:
	// Move the image's center to the screen's upper-left corner.
	// This is a preparation for rotating. When geometry matrices are applied,
	// the origin point is the upper-left corner.
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(s.theta)
	op.GeoM.Translate(float64(s.x), float64(s.y))
	screen.DrawImage(s.img, op)
}

type Game struct {
	fullscreen  bool
	s           []*Shape
	activeShape int
}

func (g *Game) Update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.s[g.activeShape].MoveBy(0, -translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.s[g.activeShape].MoveBy(0, translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.s[g.activeShape].MoveBy(-translateFactor, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.s[g.activeShape].MoveBy(translateFactor, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.s[g.activeShape].theta -= rotateFactor
	}

	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.s[g.activeShape].theta += rotateFactor
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.activeShape = (g.activeShape + 1) % len(g.s)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Because we draw in slice order, the latest is the one on top,
		// so check from latest to first
		for i := len(g.s) - 1; i >= 0; i-- {
			s := g.s[i]
			if s.In(cx, cy) {
				g.activeShape = i

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
	ebitenutil.DebugPrint(screen, "Active shape: "+g.s[g.activeShape].id)

	for _, s := range g.s {
		s.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{
		s: []*Shape{
			NewShape("Triangle", 50, 50, 0, genPolygon(3, 30, color.White)),
			NewShape("Pentagon", 100, 100, 0, genPolygon(5, 30, color.RGBA{0xff, 0, 0, 0xff})),
			NewShape("Rectangle", 200, 200, 0, genRectangle(30, 30, color.RGBA{0xff, 0, 0, 0xff})),
			NewShape("Circle", 300, 300, 0, genCircle(30, color.RGBA{0, 0xff, 0, 0xff})),
		},
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Shapes gg")

	if err := ebiten.RunGame(g); err != nil {
		if errors.Is(err, ErrCleanExit) {
			fmt.Println("Good bye!")

			return
		}

		log.Fatal(err)
	}
}
