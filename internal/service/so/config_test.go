package so

import (
	"context"
	"testing"
)

type fakeConfigStore struct {
	version int
	arm     string
	channel string
	value   string
}

func (s *fakeConfigStore) FindValue(_ context.Context, version int, arm string, channel string) (string, error) {
	s.version = version
	s.arm = arm
	s.channel = channel
	return s.value, nil
}

func TestConfigServiceListDecodesValue(t *testing.T) {
	store := &fakeConfigStore{value: `{"channel":"xj","version":511,"isUpdate":true}`}
	service := NewConfigService(store)

	data, err := service.List(context.Background(), 510, "v8a", "xj")
	if err != nil {
		t.Fatalf("list so config: %v", err)
	}

	row, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object data, got %T", data.Data)
	}
	if row["channel"] != "xj" {
		t.Fatalf("unexpected channel %v", row["channel"])
	}
	if row["version"] != float64(511) {
		t.Fatalf("unexpected version %v", row["version"])
	}
	if row["isUpdate"] != true {
		t.Fatalf("unexpected isUpdate %v", row["isUpdate"])
	}
}

func TestConfigServiceListReturnsNilForMissingOrInvalidJSON(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "missing", value: ""},
		{name: "invalid", value: "{bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewConfigService(&fakeConfigStore{value: tt.value})

			data, err := service.List(context.Background(), 0, "", "")
			if err != nil {
				t.Fatalf("list so config: %v", err)
			}
			if data.Data != nil {
				t.Fatalf("expected nil data, got %#v", data.Data)
			}
		})
	}
}

func TestConfigServiceListSanitizesLegacyStringInputs(t *testing.T) {
	store := &fakeConfigStore{value: "{}"}
	service := NewConfigService(store)

	_, err := service.List(context.Background(), 1, "<v8a>\x00", "<xj>")
	if err != nil {
		t.Fatalf("list so config: %v", err)
	}
	if store.arm != "v8a" {
		t.Fatalf("unexpected arm %q", store.arm)
	}
	if store.channel != "xj" {
		t.Fatalf("unexpected channel %q", store.channel)
	}
}
