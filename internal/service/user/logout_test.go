package user

import (
	"context"
	"testing"
)

type fakeLogoutStore struct {
	sid string
}

func (s *fakeLogoutStore) Logout(_ context.Context, sid string) error {
	s.sid = sid
	return nil
}

func TestLogoutCleansHexToken(t *testing.T) {
	store := &fakeLogoutStore{}
	service := NewLogoutService(store)

	err := service.Logout(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if store.sid != "250f790ba71ec2b9d3855f424db2259e" {
		t.Fatalf("unexpected sid %q", store.sid)
	}
}
