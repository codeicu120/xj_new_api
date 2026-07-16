package user

import (
	"context"
	"testing"
)

type fakeRedisHashStore struct {
	key    string
	field  string
	exists bool
}

func (s *fakeRedisHashStore) HExists(_ context.Context, key string, field string) (bool, error) {
	s.key = key
	s.field = field
	return s.exists, nil
}

func TestAccountDeletionExistsUsesDelAccountListHash(t *testing.T) {
	redis := &fakeRedisHashStore{exists: true}
	repo := NewRepository(nil).WithDeletionList(redis)

	exists, err := repo.AccountDeletionExists(context.Background(), 7)
	if err != nil {
		t.Fatalf("account deletion exists: %v", err)
	}
	if !exists {
		t.Fatal("expected deletion list hit")
	}
	if redis.key != "delAccountList" || redis.field != "7" {
		t.Fatalf("unexpected redis lookup key=%q field=%q", redis.key, redis.field)
	}
}
