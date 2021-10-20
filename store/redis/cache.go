package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type CacheInterface interface {
	SetWithTimout(ctx context.Context, key, val string, timeout time.Duration) error
	Unlink(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
}


type CacheStore struct {
	client redis.Cmdable
}

func (c CacheStore) SetWithTimout(ctx context.Context, key, val string, timeout time.Duration) error {
	return c.client.Set(ctx, key, val, timeout).Err()
}

func (c CacheStore) Unlink(ctx context.Context, key string) error {
	return c.client.Unlink(ctx, key).Err()
}

func (c CacheStore) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx,key).Result()
}

func NewRedisCacheStore(db redis.Cmdable) CacheInterface {
	return CacheStore{client: db}
}