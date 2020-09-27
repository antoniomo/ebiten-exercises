package main

import (
	"errors"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	translate    = 1
	blocks       = 50
)

var (
	ErrCleanExit = errors.New("clean exit, no error")
	//nolint:gochecknoglobal
	emptyImage    *ebiten.Image
	selectedColor = color.RGBA{0, 0xff, 0, 0xff}
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

type Block struct {
	id   string
	x    int
	y    int
	size int
	clr  color.Color
	img  *ebiten.Image
}

func NewBlock(id, x, y, size int, clr color.Color) *Block {
	b := &Block{
		id:   strconv.Itoa(id),
		x:    x,
		y:    y,
		size: size,
		clr:  clr,
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(size), float64(size))
	op.ColorM.Scale(colorScale(clr))

	b.img, _ = ebiten.NewImage(size, size, ebiten.FilterDefault)
	_ = b.img.DrawImage(emptyImage, op)

	return b
}

// In is from the ebiten drag and drop (drag) example.
func (b *Block) In(x, y int) bool {
	// Rectangle approach, good enough here
	if x >= b.x && x <= b.x+b.size &&
		y >= b.y && y <= b.y+b.size {
		return true
	}
	return false
}

// Move moves the block by (x, y).
func (b *Block) Move(x, y int) {
	b.x += x
	b.y += y

	if b.x+b.size > screenWidth {
		b.x = screenWidth - b.size
	}

	if b.x < 0 {
		b.x = 0
	}

	if b.y+b.size > screenHeight {
		b.y = screenHeight - b.size
	}

	if b.y < 0 {
		b.y = 0
	}
}

func (b *Block) Draw(screen *ebiten.Image, clr color.Color) {
	if clr == nil {
		clr = b.clr
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.x), float64(b.y))
	op.ColorM.Scale(colorScale(clr))
	_ = screen.DrawImage(b.img, op)
}

type connected struct {
	blk1 int
	blk2 int
}

type Game struct {
	fullscreen  bool
	blocks      []*Block
	connections []connected
	selected    int
}

func (g *Game) Update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.blocks[g.selected].Move(0, -translate)
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.blocks[g.selected].Move(0, translate)
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.blocks[g.selected].Move(-translate, 0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.blocks[g.selected].Move(translate, 0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Because we draw in slice order, the latest is the one on top,
		// so check from latest to first
		for i := len(g.blocks) - 1; i >= 0; i-- {
			b := g.blocks[i]
			if b.In(cx, cy) {
				g.selected = i

				break
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		cx, cy := ebiten.CursorPosition()
		// Because we draw in slice order, the latest is the one on top,
		// so check from latest to first
		for i := len(g.blocks) - 1; i >= 0; i-- {
			b := g.blocks[i]
			if b.In(cx, cy) {
				if i != g.selected {
					g.connect(g.selected, i)
				}

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
	ebitenutil.DebugPrint(screen, "Active block: "+g.blocks[g.selected].id)

	// Draw connections first
	for _, c := range g.connections {
		b1 := g.blocks[c.blk1]
		b2 := g.blocks[c.blk2]
		b1x := float64(b1.x + b1.size/2)
		b1y := float64(b1.y + b1.size/2)
		b2x := float64(b2.x + b2.size/2)
		b2y := float64(b2.y + b2.size/2)
		ebitenutil.DrawLine(screen, b1x, b1y, b2x, b2y, color.White)
	}

	for i, b := range g.blocks {
		if i == g.selected {
			b.Draw(screen, selectedColor)
		} else {
			b.Draw(screen, nil)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func (g *Game) init() {
	// x and y coordinates, randomized
	xs := rand.Perm(screenWidth)[:blocks]
	ys := rand.Perm(screenHeight)[:blocks]

	g.blocks = make([]*Block, blocks)
	for i, x := range xs {
		g.blocks[i] = NewBlock(i, x, ys[i], 3, color.White)
	}
}

func (g *Game) connect(blk1, blk2 int) {
	g.connections = append(g.connections, connected{blk1, blk2})
}

func main() {
	g := &Game{}
	g.init()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Connect Lines")

	if err := ebiten.RunGame(g); err != nil {
		if errors.Is(err, ErrCleanExit) {
			fmt.Println("Good bye!")

			return
		}

		log.Fatal(err)
	}
}
