package payment

import (
	"context"
	"testing"
)

func TestUnpaidAlwaysReturnsZeroTotal(t *testing.T) {
	service := NewService()

	data := service.Unpaid(context.Background())

	if data["total_count"] != 0 {
		t.Fatalf("expected total_count 0, got %v", data["total_count"])
	}
	if len(data) != 1 {
		t.Fatalf("expected only total_count, got %#v", data)
	}
}

func TestCallbackMessages(t *testing.T) {
	service := NewService()

	if got := service.SuccessMessage(context.Background()); got != "支付成功回调" {
		t.Fatalf("unexpected success message %q", got)
	}
	if got := service.FailedMessage(context.Background()); got != "支付失败回调" {
		t.Fatalf("unexpected failed message %q", got)
	}
}
