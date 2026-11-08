// Command tileset draws the course's own 16×16 tileset (lesson 28) with
// nothing but the standard library. The result is released under CC0 – use
// it for anything. Run from the repository root:
//
//	go run tools/tileset/main.go lessons/28-rust-grafik/scaffold/client/assets/tiles.png
//
// The order of the tiles must match tile_sprite and entity_sprite in the
// client (tileset.rs).
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand/v2"
	"os"
)

const T = 16 // pixels per tile

// The tiles, in atlas order.
var tiles = []func(img *image.NRGBA, ox int){
	floor, wall, grass, tree, water, rock, door, stairsUp, stairsDown,
	person(color.NRGBA{250, 220, 60, 255}), // 9: you
	person(color.NRGBA{90, 210, 250, 255}), // 10: another player
	creature,                               // 11
	item,                                   // 12
	unknown,                                // 13
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run tools/tileset/main.go <out.png>")
	}
	img := image.NewNRGBA(image.Rect(0, 0, T*16, T))
	for i, draw := range tiles {
		draw(img, i*T)
	}
	f, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 255} }

func fill(img *image.NRGBA, ox int, c color.NRGBA) {
	for y := 0; y < T; y++ {
		for x := 0; x < T; x++ {
			img.SetNRGBA(ox+x, y, c)
		}
	}
}

func px(img *image.NRGBA, ox, x, y int, c color.NRGBA) {
	if x >= 0 && y >= 0 && x < T && y < T {
		img.SetNRGBA(ox+x, y, c)
	}
}

// speckle sprinkles n pixels of c, always the same ones for the same seed.
func speckle(img *image.NRGBA, ox int, seed uint64, n int, c color.NRGBA) {
	r := rand.New(rand.NewPCG(seed, 7))
	for i := 0; i < n; i++ {
		px(img, ox, r.IntN(T), r.IntN(T), c)
	}
}

func floor(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(58, 50, 44))
	speckle(img, ox, 1, 14, rgb(80, 70, 60))
}

func wall(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(40, 38, 40))
	for row := 0; row < 4; row++ {
		shift := (row % 2) * 4
		for x := 0; x < T; x++ {
			for y := row*4 + 1; y < row*4+4; y++ {
				if (x+shift)%8 != 0 {
					px(img, ox, x, y, rgb(120, 112, 108))
				}
			}
		}
	}
}

func grass(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(40, 90, 40))
	r := rand.New(rand.NewPCG(3, 3))
	for i := 0; i < 9; i++ {
		x, y := r.IntN(T), 3+r.IntN(T-3)
		px(img, ox, x, y, rgb(110, 190, 80))
		px(img, ox, x, y-1, rgb(110, 190, 80))
	}
}

func tree(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(30, 70, 30))
	for y := 11; y < 15; y++ {
		px(img, ox, 7, y, rgb(110, 75, 40))
		px(img, ox, 8, y, rgb(110, 75, 40))
	}
	for y := 1; y < 12; y++ {
		for x := 2; x < 14; x++ {
			if dx, dy := float64(x)-7.5, float64(y)-6; dx*dx+dy*dy*1.2 < 30 {
				c := rgb(50, 150, 60)
				if dx+dy < -3 {
					c = rgb(90, 190, 80)
				}
				px(img, ox, x, y, c)
			}
		}
	}
}

func water(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(20, 50, 120))
	for _, y0 := range []int{3, 8, 13} {
		for x := 0; x < T; x++ {
			y := y0 + int(math.Round(math.Sin(float64(x)/2.5)))
			px(img, ox, x, y, rgb(90, 150, 230))
		}
	}
}

func rock(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(50, 48, 52))
	for y := 2; y < 15; y++ {
		half := (y - 2) / 2
		for x := 8 - half; x <= 8+half; x++ {
			c := rgb(140, 140, 150)
			if x < 8 {
				c = rgb(175, 175, 185)
			}
			px(img, ox, x, y, c)
		}
	}
}

func door(img *image.NRGBA, ox int) {
	fill(img, ox, rgb(58, 50, 44))
	for y := 1; y < 16; y++ {
		for x := 3; x < 13; x++ {
			c := rgb(140, 95, 50)
			if x%3 == 0 {
				c = rgb(100, 65, 35)
			}
			px(img, ox, x, y, c)
		}
	}
	px(img, ox, 10, 9, rgb(230, 200, 90))
}

func steps(img *image.NRGBA, ox int, up bool) {
	fill(img, ox, rgb(58, 50, 44))
	for i := 0; i < 4; i++ {
		y := 3 + i*3
		x0 := 2 + i*2
		if !up {
			x0 = 2 + (3-i)*2
		}
		for x := x0; x < 14; x++ {
			px(img, ox, x, y, rgb(200, 200, 200))
			px(img, ox, x, y+1, rgb(130, 130, 130))
		}
	}
}

func stairsUp(img *image.NRGBA, ox int)   { steps(img, ox, true) }
func stairsDown(img *image.NRGBA, ox int) { steps(img, ox, false) }

// person draws a little figure on a transparent background.
func person(c color.NRGBA) func(*image.NRGBA, int) {
	return func(img *image.NRGBA, ox int) {
		for y := 1; y < 5; y++ {
			for x := 6; x < 10; x++ {
				px(img, ox, x, y, rgb(240, 200, 160))
			}
		}
		for y := 5; y < 11; y++ {
			for x := 4; x < 12; x++ {
				px(img, ox, x, y, c)
			}
		}
		for y := 11; y < 15; y++ {
			px(img, ox, 5, y, rgb(60, 60, 80))
			px(img, ox, 6, y, rgb(60, 60, 80))
			px(img, ox, 9, y, rgb(60, 60, 80))
			px(img, ox, 10, y, rgb(60, 60, 80))
		}
	}
}

func creature(img *image.NRGBA, ox int) {
	body := rgb(200, 70, 60)
	for y := 6; y < 12; y++ {
		for x := 2; x < 13; x++ {
			px(img, ox, x, y, body)
		}
	}
	for y := 3; y < 8; y++ {
		for x := 10; x < 15; x++ {
			px(img, ox, x, y, body)
		}
	}
	px(img, ox, 13, 4, rgb(255, 240, 80))
	for _, x := range []int{3, 5, 9, 11} {
		px(img, ox, x, 12, body)
		px(img, ox, x, 13, body)
	}
	px(img, ox, 1, 6, body)
	px(img, ox, 0, 5, body)
}

func item(img *image.NRGBA, ox int) {
	for y := 6; y < 13; y++ {
		for x := 3; x < 13; x++ {
			c := rgb(160, 110, 40)
			if y == 8 {
				c = rgb(230, 190, 70)
			}
			px(img, ox, x, y, c)
		}
	}
	px(img, ox, 8, 9, rgb(250, 240, 150))
}

func unknown(img *image.NRGBA, ox int) {
	c := rgb(255, 0, 255)
	for _, p := range [][2]int{{6, 3}, {7, 2}, {8, 2}, {9, 3}, {10, 4}, {10, 5}, {9, 6}, {8, 7}, {8, 8}, {8, 9}, {8, 12}} {
		px(img, ox, p[0], p[1], c)
		px(img, ox, p[0]+1, p[1], c)
	}
}
