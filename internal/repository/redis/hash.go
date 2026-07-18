package redis

import (
	"context"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type HashStore struct {
	client *goredis.Client
}

func NewHashStore(addr string, password string, db int) *HashStore {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}
	return &HashStore{client: goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})}
}

func (s *HashStore) HExists(ctx context.Context, key string, field string) (bool, error) {
	if s == nil || s.client == nil {
		return false, nil
	}
	return s.client.HExists(ctx, key, field).Result()
}

func (s *HashStore) HSet(ctx context.Context, key string, field string, value interface{}) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.HSet(ctx, key, field, value).Err()
}

func (s *HashStore) HDel(ctx context.Context, key string, field string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.HDel(ctx, key, field).Err()
}
