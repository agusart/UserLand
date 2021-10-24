package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type SessionCache struct {
	Id uint `redis:"id"`
	UserId uint `redis:"user_id"`
	JwtId string `redis:"jwt_id"`
}

type CacheInterface interface {
	SetWithTimout(ctx context.Context, key, val string, timeout time.Duration) error
	Unlink(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	InsertSessionCache(ctx context.Context, cache SessionCache) error
	GetSessionCache(ctx context.Context, userId, sessionId uint) (*SessionCache, error)
	DeleteSessionCache(ctx context.Context, userId, sessionId uint) error
}


type CacheStore struct {
	client redis.Cmdable
}

func (c CacheStore) DeleteSessionCache(ctx context.Context, userId, sessionId uint) error {
	key := GenerateUserSessionKey(userId, sessionId)
	return c.Unlink(ctx, key)
}

func (c CacheStore) InsertSessionCache(ctx context.Context, cache SessionCache) error {
	key := GenerateUserSessionKey(cache.UserId, cache.Id)
	args := StructToArgs(cache)
	err := c.client.HMSet(ctx, key, args...).Err()

	return err
}

func (c CacheStore) GetSessionCache(ctx context.Context, userId, sessionId uint) (*SessionCache, error) {
	key := GenerateUserSessionKey(userId, sessionId)
	sessionCache := SessionCache{}

	err := c.client.HGetAll(ctx, key).Scan(&sessionCache)
	if err != nil {
		return nil, err
	}

	return &sessionCache, nil
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