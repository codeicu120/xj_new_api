package captcha

import (
	"bytes"
	"crypto/rand"
	"image"
	"image/color"
	"image/png"
	"math/big"
)

const (
	testImageWidth  = 100
	testImageHeight = 34
)

type TestImageService struct{}

func NewTestImageService() *TestImageService {
	return &TestImageService{}
}

func (s *TestImageService) PNG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, testImageWidth, testImageHeight))
	fill(img, color.RGBA{R: 246, G: 249, B: 242, A: 255})
	drawNoise(img, 120)
	drawLine(img, 3, 5, 96, 28, color.RGBA{R: 140, G: 155, B: 185, A: 255})
	drawLine(img, 0, 23, 99, 8, color.RGBA{R: 180, G: 120, B: 135, A: 255})

	drawGlyph(img, 17, 6, randSeed(), color.RGBA{R: 42, G: 64, B: 93, A: 255})
	drawGlyph(img, 56, 5, randSeed(), color.RGBA{R: 94, G: 54, B: 74, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fill(img *image.RGBA, c color.RGBA) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func drawNoise(img *image.RGBA, n int) {
	for i := 0; i < n; i++ {
		x := randInt(testImageWidth)
		y := randInt(testImageHeight)
		img.SetRGBA(x, y, color.RGBA{
			R: uint8(120 + randInt(100)),
			G: uint8(120 + randInt(100)),
			B: uint8(120 + randInt(100)),
			A: 255,
		})
	}
}

func drawGlyph(img *image.RGBA, ox, oy, seed int, c color.RGBA) {
	for row := 0; row < 7; row++ {
		for col := 0; col < 6; col++ {
			bit := (seed >> ((row + col) % 15)) & 1
			border := row == 0 || row == 6 || col == 0 || col == 5
			if bit == 1 || border && (row+col+seed)%3 == 0 {
				drawBlock(img, ox+col*4, oy+row*3, 3, 3, c)
			}
		}
	}
}

func drawBlock(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if image.Pt(xx, yy).In(img.Bounds()) {
				img.SetRGBA(xx, yy, c)
			}
		}
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		drawBlock(img, x0, y0, 2, 2, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func randSeed() int {
	return randInt(1<<15) + 1
}

func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
