package minivod

import (
	"context"
	"testing"
)

func TestLongToShortMapByLongIDNilDB(t *testing.T) {
	repo := NewRepository(nil)

	row, err := repo.LongToShortMapByLongID(context.Background(), 9)
	if err != nil {
		t.Fatalf("LongToShortMapByLongID: %v", err)
	}
	if len(row) != 0 {
		t.Fatalf("expected empty row, got %#v", row)
	}
}
