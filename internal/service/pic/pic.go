package pic

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidRequest = errors.New("invalid pic request")
	ErrNotFound       = errors.New("pic not found")
)

type Service struct {
	uploadPath string
}

var sizes = map[string]image.Point{
	"C1": {100, 100}, "C2": {200, 200}, "C3": {300, 300}, "C4": {400, 400}, "C5": {500, 500}, "C6": {600, 600}, "C7": {700, 700}, "C8": {800, 800}, "C9": {900, 900},
	"T1": {100, 0}, "T2": {200, 0}, "T3": {300, 0}, "T4": {400, 0}, "T5": {500, 0}, "T6": {600, 0}, "T7": {700, 0}, "T8": {800, 0}, "T9": {900, 0},
	"R1": {100, 100}, "R2": {200, 200}, "R3": {300, 300}, "R4": {400, 400}, "R5": {500, 500}, "R6": {600, 600}, "R7": {700, 700}, "R8": {800, 800}, "R9": {900, 900},
	"M": {0, 0}, "N": {0, 0},
}

func NewService(uploadPath string) *Service {
	return &Service{uploadPath: uploadPath}
}

func (s *Service) Image(_ context.Context, size string, uri string) ([]byte, string, error) {
	if _, ok := sizes[size]; !ok || !validImageURI(uri) {
		return nil, "", ErrInvalidRequest
	}
	path, err := s.absPath(uri)
	if err != nil {
		return nil, "", ErrInvalidRequest
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	contentType := contentTypeForURI(uri)
	if size == "N" || size == "M" {
		return raw, contentType, nil
	}

	src, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", err
	}
	var dst image.Image
	if strings.HasPrefix(size, "C") {
		dst = cropResize(src, sizes[size].X, sizes[size].Y)
	} else {
		dst = thumb(src, sizes[size].X, sizes[size].Y)
	}
	encoded, err := encode(dst, format)
	if err != nil {
		return nil, "", err
	}
	return encoded, contentTypeForFormat(format, contentType), nil
}

func (s *Service) absPath(uri string) (string, error) {
	clean := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(uri, "\\", "/"), "/"))
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", ErrInvalidRequest
	}
	return filepath.Join(s.uploadPath, clean), nil
}

func validImageURI(uri string) bool {
	lower := strings.ToLower(uri)
	return strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".png")
}

func contentTypeForURI(uri string) string {
	lower := strings.ToLower(uri)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

func contentTypeForFormat(format string, fallback string) string {
	switch format {
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "jpeg":
		return "image/jpeg"
	default:
		return fallback
	}
}

func cropResize(src image.Image, width int, height int) image.Image {
	bounds := src.Bounds()
	sw, sh := bounds.Dx(), bounds.Dy()
	srcRatio := float64(sw) / float64(sh)
	dstRatio := float64(width) / float64(height)
	cw, ch := sw, sh
	if dstRatio > srcRatio {
		ch = int(float64(sw) / dstRatio)
	} else {
		cw = int(float64(sh) * dstRatio)
	}
	cx := bounds.Min.X + (sw-cw)/2
	cy := bounds.Min.Y + (sh-ch)/2
	return resizeFromRect(src, image.Rect(cx, cy, cx+cw, cy+ch), width, height, false)
}

func thumb(src image.Image, width int, height int) image.Image {
	bounds := src.Bounds()
	sw, sh := bounds.Dx(), bounds.Dy()
	if height == 0 {
		if width > sw {
			width = sw
		}
		height = int(float64(width) / (float64(sw) / float64(sh)))
		return resizeFromRect(src, bounds, width, height, false)
	}
	ratioSrc := float64(sw) / float64(sh)
	ratioDst := float64(width) / float64(height)
	dw, dh := width, height
	if ratioDst > ratioSrc {
		dh = height
		dw = int(float64(height) * ratioSrc)
	} else {
		dw = width
		dh = int(float64(width) / ratioSrc)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(canvas, color.White)
	resized := resizeFromRect(src, bounds, dw, dh, false)
	x := (width - dw) / 2
	y := (height - dh) / 2
	copyImage(canvas, resized, x, y)
	return canvas
}

func resizeFromRect(src image.Image, rect image.Rectangle, width int, height int, white bool) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if white {
		fill(dst, color.White)
	}
	for y := 0; y < height; y++ {
		sy := rect.Min.Y + y*rect.Dy()/height
		for x := 0; x < width; x++ {
			sx := rect.Min.X + x*rect.Dx()/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func fill(dst *image.RGBA, c color.Color) {
	for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
		for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
			dst.Set(x, y, c)
		}
	}
}

func copyImage(dst *image.RGBA, src image.Image, offX int, offY int) {
	b := src.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(offX+x, offY+y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
}

func encode(img image.Image, format string) ([]byte, error) {
	buf := bytes.Buffer{}
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "gif":
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, err
		}
	default:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
