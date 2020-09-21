package main

import (
	"errors"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	// Very simplistic 2-layer parallax. Of course this could be fully
	// dynamic with translate speed based on distance, and each star it's
	// own distance...
	translateNear = 3
	translateFar  = 1
	nearStars     = 50
	farStars      = 100
)

var (
	ErrCleanExit = errors.New("clean exit, no error")
	//nolint:gochecknoglobal
	emptyImage *ebiten.Image
)

//nolint:gochecknoinit
func init() {
	rand.Seed(time.Now().UnixNano())

	emptyImage, _ = ebiten.NewImage(1, 1, ebiten.FilterDefault)
	_ = emptyImage.Fill(color.White)
}

// colorScale taken from ebitenutil/shapes.go.
func colorScale(clr color.Color) (rf, gf, bf, af float64) {
	r, g, b, a := clr.RGBA()
	if a == 0 {
		return 0, 0, 0, 0
	}

	rf = float64(r) / float64(a)
	gf = float64(g) / float64(a)
	bf = float64(b) / float64(a)
	af = float64(a) / 0xffff

	return
}

type Star struct {
	x      int
	y      int
	radius int
	img    *ebiten.Image
}

func NewStar(x, y, radius int, clr color.Color) *Star {
	s := &Star{
		x:      x,
		y:      y,
		radius: radius,
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(radius*2), float64(radius*2))
	op.ColorM.Scale(colorScale(clr))

	s.img, _ = ebiten.NewImage(radius*2, radius*2, ebiten.FilterDefault)
	_ = s.img.DrawImage(emptyImage, op)

	return s
}

// In is from the ebiten drag and drop (drag) example.
func (s *Star) In(x, y int) bool {
	// Rectangle approach, not precise for triangles but good enough here
	// if x >= p.x-p.radius && x <= p.x+p.radius &&
	// 	y >= p.y-p.radius && y <= p.y+p.radius {
	// 	return true
	// }
	//
	// return false
	return s.img.At(x-s.x+s.radius, y-s.y+s.radius).(color.RGBA).A > 0
}

// MoveBy moves the star by (x, y).
func (s *Star) MoveBy(x, y int) {
	s.x += x
	s.y += y

	// Circular stars
	if s.x > screenWidth {
		s.x = 0
	}

	if s.x < 0 {
		s.x = screenWidth
	}

	if s.y > screenHeight {
		s.y = 0
	}

	if s.y < 0 {
		s.y = screenHeight
	}
}

func (s *Star) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x), float64(s.y))
	_ = screen.DrawImage(s.img, op)
}

type Game struct {
	fullscreen bool
	nearStars  []*Star
	farStars   []*Star
}

func (g *Game) MoveView(x, y int) {
	for _, s := range g.nearStars {
		s.MoveBy(x*translateNear, y*translateNear)
	}
	for _, s := range g.farStars {
		s.MoveBy(x*translateFar, y*translateFar)
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.MoveView(0, -1)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.MoveView(0, 1)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.MoveView(-1, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.MoveView(1, 0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrCleanExit
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, s := range g.nearStars {
		s.Draw(screen)
	}
	for _, s := range g.farStars {
		s.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func (g *Game) initStarfield() {
	// NewStar(3, 3, 5, color.White)
	// NewStar(15, 15, 10, color.RGBA{0xff, 0, 0, 0xff})
	// NewStar(100, 100, 15, color.RGBA{0, 0xff, 0, 0xff})

	// x and y coordinates, randomized
	xs := rand.Perm(screenWidth)[:nearStars]
	ys := rand.Perm(screenHeight)[:nearStars]

	g.nearStars = make([]*Star, nearStars)
	for i, x := range xs {
		g.nearStars[i] = NewStar(x, ys[i], 3, color.White)
	}

	// x and y coordinates, randomized
	xs = rand.Perm(screenWidth)[:farStars]
	ys = rand.Perm(screenHeight)[:farStars]

	g.farStars = make([]*Star, farStars)
	for i, x := range xs {
		// Dim farther stars with alpha channel
		g.farStars[i] = NewStar(x, ys[i], 3, color.RGBA{0xff, 0xff, 0xff, 0x80})
	}
}

func main() {
	g := &Game{}
	g.initStarfield()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Starfield")

	if err := ebiten.RunGame(g); err != nil {
		if errors.Is(err, ErrCleanExit) {
			fmt.Println("Good bye!")

			return
		}

		log.Fatal(err)
	}
}
