package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strings"

	"path/filepath"

	"github.com/StephaneBunel/bresenham"
	"github.com/kmicki/apng"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// constants
const CELL_SIZE = 60

// variables for color
var (
	green     = color.RGBA{A: 255, G: 255}
	darkGreen = color.RGBA{R: 1, G: 100, B: 32, A: 255}
	red       = color.RGBA{R: 255, A: 255}
	yellow    = color.RGBA{R: 255, G: 255, B: 101, A: 255}
	gray      = color.RGBA{R: 125, G: 125, B: 125, A: 125}
	orange    = color.RGBA{R: 255, G: 140, B: 25, A: 255}
	blue      = color.RGBA{R: 14, G: 118, B: 173, A: 255}
)

// draw to png
func (g *Maze) OutputImage(fileName ...string) {
	width := CELL_SIZE * g.Width
	height := CELL_SIZE * g.Height

	fmt.Printf("Generating image (%dx%d)...\n", width, height)

	name := "image.png"
	if len(fileName) > 0 {
		name = fileName[0]
	}

	outFile := filepath.Join(g.OutputDir, name)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	for i, row := range g.Walls {
		for j, col := range row {
			p := Point{Row: i, Col: j}
			var cellColor color.Color

			switch {
			// first priority: start and goal
			case p == g.Start:
				cellColor = darkGreen
			case p == g.Goal:
				cellColor = red

			// second priority: walls
			case col.wall:
				cellColor = color.Black

			// third priority: search state -> p isnt start / goal
			case p == g.CurrentNode.State:
				cellColor = orange

			// fourth priority: is this cell flooded?
			case col.State.Water:
				cellColor = blue

			case g.inSolution(p):
				cellColor = green
			case inExplored(p, g.Explored):
				cellColor = yellow

			// empty space
			default:
				cellColor = color.White
			}

			g.drawSquare(col, p, img, cellColor, j*CELL_SIZE, i*CELL_SIZE)
		}
	}

	// horizontal lines
	for i := 0; i <= g.Height; i++ {
		y := i * CELL_SIZE
		// ensure we don't draw until the last pixel
		if y == height {
			y--
		}
		bresenham.DrawLine(img, 0, y, width, y, gray)
	}

	// vertical lines
	for j := 0; j <= g.Width; j++ {
		x := j * CELL_SIZE
		if x == width {
			x--
		}
		bresenham.DrawLine(img, x, 0, x, height, gray)
	}

	// save
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
	}

	fmt.Printf("Successfully generated: %s\n", outFile)
}

func (g *Maze) drawSquare(col Wall, p Point, img *image.RGBA, c color.Color, x, y int) {
	rect := image.Rect(x, y, x+CELL_SIZE, y+CELL_SIZE)
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
	if !col.wall {
		switch g.SearchType {
		case DIJKSTRA, GBFS:
			g.printManhattanCost(p, color.Black, img, x, y)
		case ASTAR:
			g.printTotalCost(p, color.Black, img, x, y)
		default:
			// do nothing
		}

		// check to see if this cell is flooded
		if col.State.Water {
			g.printWater(blue, img, x, y)
		}

		g.printLocation(p, color.Black, img, x, y)
	}
}

func (g *Maze) printWater(c color.Color, img *image.RGBA, x, y int) {
	text := "W"

	textWidth := len(text) * 7
	offsetX := (60 - textWidth) / 2
	offsetY := 25

	dot := fixed.Point26_6{
		X: fixed.I(x + offsetX),
		Y: fixed.I(y + offsetY),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  dot,
	}

	d.DrawString(text)
}

func (g *Maze) printTotalCost(p Point, c color.Color, img *image.RGBA, x, y int) {
	n := Node{State: p}

	fromStart := n.ManhattanDistance(g.Start)
	toGoal := EuclideanDistance(p, g.Goal)

	text := fmt.Sprintf("%.2f", float64(fromStart)+toGoal)

	textWidth := len(text) * 7
	offsetX := (60 - textWidth) / 2
	offsetY := 25

	dot := fixed.Point26_6{
		X: fixed.I(x + offsetX),
		Y: fixed.I(y + offsetY),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  dot,
	}

	d.DrawString(text)
}

func (g *Maze) printManhattanCost(p Point, c color.Color, img *image.RGBA, x, y int) {
	n := Node{State: p}

	var manhattanDistance int
	switch g.SearchType {
	case DIJKSTRA:
		manhattanDistance = n.ManhattanDistance(g.Start)
	case GBFS:
		manhattanDistance = n.ManhattanDistance(g.Goal)
	default:
		// nothing here
	}

	text := fmt.Sprintf("%d", manhattanDistance)

	textWidth := len(text) * 7
	offsetX := (60 - textWidth) / 2
	offsetY := 25

	dot := fixed.Point26_6{
		X: fixed.I(x + offsetX),
		Y: fixed.I(y + offsetY),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  dot,
	}

	d.DrawString(text)
}

func (g *Maze) printLocation(p Point, c color.Color, img *image.RGBA, offsetX, offsetY int) {
	dot := fixed.Point26_6{
		X: fixed.I(offsetX + 6),
		Y: fixed.I(offsetY + 40),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  dot,
	}

	d.DrawString(fmt.Sprintf("[%d,%d]", p.Row, p.Col))
}

func (g *Maze) OutputAnimateImage() {
	output := filepath.Join(g.OutputDir, "animation.png")
	frameDir := g.FrameDir
	files, err := os.ReadDir(frameDir)
	if err != nil {
		log.Printf("Could not read frames directory: %v\n", err)
		return
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			images = append(images, filepath.Join(frameDir, file.Name()))
		}
	}

	images = append(images, filepath.Join(g.OutputDir, "image.png"))

	a := apng.APNG{Frames: make([]apng.Frame, len(images))}

	out, err := os.Create(output)
	if err != nil {
		log.Printf("Could not create animation file: %v\n", err)
		return
	}
	defer out.Close()

	fmt.Printf("Encoding %d frames into animation...\n", len(images))

	for i, s := range images {
		in, err := os.Open(s)
		if err != nil {
			log.Printf("Error opening frame %s: %v\n", s, err)
			continue
		}

		m, err := png.Decode(in)
		in.Close()

		if err != nil {
			log.Printf("Error decoding %s: %v\n", s, err)
			continue
		}

		a.Frames[i].Image = m
	}

	if err := apng.Encode(out, a); err != nil {
		log.Printf("Error encoding APNG: %v\n", err)
		return
	}

	fmt.Printf("Animation saved to: %s\n", output)
}
