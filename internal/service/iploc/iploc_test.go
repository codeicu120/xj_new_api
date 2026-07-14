package iploc

import (
	"errors"
	"testing"
)

type fakeLocator struct {
	parts []string
	err   error
}

func (l fakeLocator) Find(string) ([]string, error) {
	return l.parts, l.err
}

func TestServiceFind(t *testing.T) {
	service := NewService(fakeLocator{parts: []string{"GOOGLE.COM", "GOOGLE.COM"}})

	data := service.Find("8.8.8.8")

	if data.Data != "GOOGLE.COM GOOGLE.COM" {
		t.Fatalf("unexpected data %q", data.Data)
	}
}

func TestServiceFindError(t *testing.T) {
	service := NewService(fakeLocator{err: errors.New("not found")})

	data := service.Find("bad")

	if data.Data != "" {
		t.Fatalf("expected empty data, got %q", data.Data)
	}
}
