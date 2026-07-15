package pic

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestImageRejectsInvalidRequest(t *testing.T) {
	service := NewService(t.TempDir())

	_, _, err := service.Image(context.Background(), "X1", "a.png")
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}

	_, _, err = service.Image(context.Background(), "C1", "a.txt")
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestImageReturnsNotFound(t *testing.T) {
	service := NewService(t.TempDir())

	_, _, err := service.Image(context.Background(), "C1", "missing.png")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestImageOriginalAndResize(t *testing.T) {
	dir := t.TempDir()
	source := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			source.Set(x, y, color.RGBA{R: uint8(x * 5), G: uint8(y * 10), B: 80, A: 255})
		}
	}
	buf := bytes.Buffer{}
	if err := png.Encode(&buf, source); err != nil {
		t.Fatalf("encode source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample.png"), buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	service := NewService(dir)
	raw, contentType, err := service.Image(context.Background(), "N", "sample.png")
	if err != nil {
		t.Fatalf("original image: %v", err)
	}
	if contentType != "image/png" || !bytes.Equal(raw, buf.Bytes()) {
		t.Fatalf("unexpected original contentType=%q equal=%v", contentType, bytes.Equal(raw, buf.Bytes()))
	}

	cropped, contentType, err := service.Image(context.Background(), "C1", "sample.png")
	if err != nil {
		t.Fatalf("crop image: %v", err)
	}
	assertPNGSize(t, cropped, contentType, 100, 100)

	thumbed, contentType, err := service.Image(context.Background(), "T1", "sample.png")
	if err != nil {
		t.Fatalf("thumb image: %v", err)
	}
	assertPNGSize(t, thumbed, contentType, 40, 20)
}

func assertPNGSize(t *testing.T, raw []byte, contentType string, width int, height int) {
	t.Helper()
	if contentType != "image/png" {
		t.Fatalf("expected image/png, got %q", contentType)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
		t.Fatalf("expected %dx%d, got %dx%d", width, height, img.Bounds().Dx(), img.Bounds().Dy())
	}
}
