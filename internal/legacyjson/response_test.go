package legacyjson

import (
	"encoding/json"
	"testing"
)

func TestOKResponseShape(t *testing.T) {
	response := OK(map[string]string{"picurl": "/captcha/pic?secret"})

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["retcode"] != float64(0) {
		t.Fatalf("expected retcode 0, got %v", body["retcode"])
	}
	if body["errmsg"] != "" {
		t.Fatalf("expected empty errmsg, got %q", body["errmsg"])
	}
	if _, ok := body["data"]; !ok {
		t.Fatal("expected data field")
	}
}

func TestErrorResponseOmitsEmptyData(t *testing.T) {
	payload, err := json.Marshal(Error("验证失败"))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["retcode"] != float64(-1) {
		t.Fatalf("expected retcode -1, got %v", body["retcode"])
	}
	if _, ok := body["data"]; ok {
		t.Fatal("expected data field to be omitted")
	}
}
