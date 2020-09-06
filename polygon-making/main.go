package main

import (
	"errors"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"

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
	emptyImage   *ebiten.Image
)

func init() {
	emptyImage, _ = ebiten.NewImage(1, 1, ebiten.FilterDefault)
	emptyImage.Fill(color.White)
}

// Based on ebiten polygons example. This is just approximate.
func genPolygon(radius, num int, color color.RGBA) ([]ebiten.Vertex, []uint16) {
	vs := make([]ebiten.Vertex, num+1)

	r := float32(color.R) / 0xff
	g := float32(color.G) / 0xff
	b := float32(color.B) / 0xff
	a := float32(color.A) / 0xff

	for i := 0; i < num; i++ {
		rate := float64(i) / float64(num)

		vs[i] = ebiten.Vertex{
			DstX:   float32(float64(radius)*math.Cos(2*math.Pi*rate)) + float32(radius),
			DstY:   float32(float64(radius)*math.Sin(2*math.Pi*rate)) + float32(radius),
			SrcX:   0,
			SrcY:   0,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
	}

	vs[len(vs)-1] = ebiten.Vertex{
		DstX:   float32(radius),
		DstY:   float32(radius),
		SrcX:   0,
		SrcY:   0,
		ColorR: r,
		ColorG: g,
		ColorB: b,
		ColorA: a,
	}

	indices := []uint16{}
	for i := 0; i < num; i++ {
		indices = append(indices, uint16(i), uint16(i+1)%uint16(num), uint16(num))
	}

	return vs, indices
}

type Polygon struct {
	id     string
	x      int
	y      int
	radius int
	theta  float64
	img    *ebiten.Image
}

func NewPolygon(id string, x, y int, theta float64, radius, sides int, clr color.RGBA) *Polygon {
	vs, indices := genPolygon(radius, sides, clr)

	p := &Polygon{
		id:     id,
		x:      x,
		y:      y,
		radius: radius,
		theta:  theta,
	}
	p.img, _ = ebiten.NewImage(radius*2, radius*2, ebiten.FilterDefault)
	p.img.DrawTriangles(vs, indices, emptyImage, nil)
	return p
}

// In is from the ebiten drag and drop (drag) example.
func (p *Polygon) In(x, y int) bool {
	// Rectangle approach, not precise for triangles but good enough here
	// if x >= p.x-p.radius && x <= p.x+p.radius &&
	// 	y >= p.y-p.radius && y <= p.y+p.radius {
	// 	return true
	// }

	// return false
	return p.img.At(x-p.x+p.radius, y-p.y+p.radius).(color.RGBA).A > 0
}

// MoveBy moves the polygon by (x, y).
func (p *Polygon) MoveBy(x, y int) {
	p.x += x
	p.y += y

	if p.x < 0+p.radius {
		p.x = 0 + p.radius
	}

	if p.x > screenWidth-p.radius {
		p.x = screenWidth - p.radius
	}

	if p.y < 0+p.radius {
		p.y = 0 + p.radius
	}

	if p.y > screenHeight-p.radius {
		p.y = screenHeight - p.radius
	}
}

func (p *Polygon) Draw(screen *ebiten.Image) {
	w, h := p.img.Size()

	op := &ebiten.DrawImageOptions{}
	// From Ebiten's rotate example:
	// Move the image's center to the screen's upper-left corner.
	// This is a preparation for rotating. When geometry matrices are applied,
	// the origin point is the upper-left corner.
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(p.theta)
	op.GeoM.Translate(float64(p.x), float64(p.y))
	screen.DrawImage(p.img, op)
}

type Game struct {
	fullscreen    bool
	p             []*Polygon
	activePolygon int
}

func (g *Game) Update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.p[g.activePolygon].MoveBy(0, -translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.p[g.activePolygon].MoveBy(0, translateFactor)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.p[g.activePolygon].MoveBy(-translateFactor, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.p[g.activePolygon].MoveBy(translateFactor, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.p[g.activePolygon].theta -= rotateFactor
	}

	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.p[g.activePolygon].theta += rotateFactor
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.activePolygon = (g.activePolygon + 1) % len(g.p)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Because we draw in slice order, the latest is the one on top,
		// so check from latest to first
		for i := len(g.p) - 1; i >= 0; i-- {
			s := g.p[i]
			if s.In(cx, cy) {
				g.activePolygon = i

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
	ebitenutil.DebugPrint(screen, "Active polygon: "+g.p[g.activePolygon].id)

	for _, p := range g.p {
		p.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{
		p: []*Polygon{
			NewPolygon("Triangle", 0, 10, 0, 20, 3, color.RGBA{0xff, 0xff, 0xff, 0xff}),
			NewPolygon("Pentagon", 50, 50, 0, 20, 5, color.RGBA{0xff, 0, 0, 0xff}),
			NewPolygon("Circle", 100, 100, 0, 20, 8, color.RGBA{0x00, 0xff, 0, 0xff}),
		},
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Polygon Making")

	if err := ebiten.RunGame(g); err != nil {
		if errors.Is(err, ErrCleanExit) {
			fmt.Println("Good bye!")

			return
		}

		log.Fatal(err)
	}
}
