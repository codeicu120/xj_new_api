package index

import (
	"context"
	"testing"
)

type fakeSettingsStore struct {
	value string
}

func (s fakeSettingsStore) SettingValue(context.Context) (string, error) {
	return s.value, nil
}

type fakeCertClient struct {
	url  string
	body []byte
	err  error
}

func (c *fakeCertClient) Get(_ context.Context, rawURL string) ([]byte, error) {
	c.url = rawURL
	return c.body, c.err
}

func TestCertServiceUsesConfiguredURLAndReturnsData(t *testing.T) {
	client := &fakeCertClient{body: []byte(`{"code":0,"data":{"uuid":"abc"}}`)}
	service := NewCertService(fakeSettingsStore{value: `a:1:{s:10:"getCertUrl";s:24:"https://cert.example/api";}`}, client)

	data, err := service.GetCertUUID(context.Background(), "id 1")
	if err != nil {
		t.Fatalf("get cert uuid: %v", err)
	}
	if client.url != "https://cert.example/api?uuid=id+1" {
		t.Fatalf("unexpected url %q", client.url)
	}
	row := data.(map[string]interface{})
	if row["uuid"] != "abc" {
		t.Fatalf("unexpected data %#v", row)
	}
}

func TestCertServiceNotFound(t *testing.T) {
	client := &fakeCertClient{body: []byte(`{"code":1,"data":null}`)}
	service := NewCertService(fakeSettingsStore{}, client)

	_, err := service.GetCertUUID(context.Background(), "missing")
	if !IsCertNotFound(err) {
		t.Fatalf("expected cert not found, got %v", err)
	}
}
