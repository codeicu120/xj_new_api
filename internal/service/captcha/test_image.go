package captcha

import (
	"bytes"
	"crypto/rand"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"strings"
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
	return s.PNGForCode("1234")
}

func (s *TestImageService) PNGForCode(code string) ([]byte, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		code = "1234"
	}
	img := image.NewRGBA(image.Rect(0, 0, testImageWidth, testImageHeight))
	fill(img, color.RGBA{R: 246, G: 249, B: 242, A: 255})
	drawNoise(img, 120)
	drawLine(img, 3, 5, 96, 28, color.RGBA{R: 140, G: 155, B: 185, A: 255})
	drawLine(img, 0, 23, 99, 8, color.RGBA{R: 180, G: 120, B: 135, A: 255})

	drawCode(img, code)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawCode(img *image.RGBA, code string) {
	runes := []rune(strings.ToUpper(code))
	if len(runes) > 4 {
		runes = runes[:4]
	}
	x := 10
	if len(runes) < 4 {
		x = 24
	}
	palette := []color.RGBA{
		{R: 60, G: 128, B: 80, A: 255},
		{R: 136, G: 58, B: 150, A: 255},
		{R: 40, G: 94, B: 160, A: 255},
		{R: 150, G: 60, B: 84, A: 255},
	}
	for i, r := range runes {
		pattern, ok := glyphPatterns[r]
		if !ok {
			pattern = glyphPatterns['0']
		}
		drawPattern(img, x+i*21, 6+randInt(5), pattern, palette[i%len(palette)])
	}
}

func drawPattern(img *image.RGBA, ox int, oy int, pattern []string, c color.RGBA) {
	for row, line := range pattern {
		for col, ch := range line {
			if ch != '1' {
				continue
			}
			drawBlock(img, ox+col*3, oy+row*3, 3, 3, c)
			if randInt(3) == 0 {
				drawBlock(img, ox+col*3+1, oy+row*3, 3, 2, color.RGBA{R: c.R / 2, G: c.G / 2, B: c.B / 2, A: 255})
			}
		}
	}
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

var glyphPatterns = map[rune][]string{
	'0': {"11111", "10001", "10011", "10101", "11001", "10001", "11111"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"11110", "00001", "00001", "11110", "10000", "10000", "11111"},
	'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4': {"10010", "10010", "10010", "11111", "00010", "00010", "00010"},
	'5': {"11111", "10000", "10000", "11110", "00001", "00001", "11110"},
	'6': {"01111", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00001", "11110"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'C': {"01111", "10000", "10000", "10000", "10000", "10000", "01111"},
	'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G': {"01111", "10000", "10000", "10011", "10001", "10001", "01111"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'J': {"00111", "00010", "00010", "00010", "10010", "10010", "01100"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
}
