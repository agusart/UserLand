package redis

import (
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
)

var (
	redisDatabase redis.Cmdable
	redisTestDatabase redis.Cmdable
)

func InitRedisDb() redis.Cmdable {
	addr := getRedisAddress()
	pass := os.Getenv("REDIS_PASSWORD")

	redisDatabase = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	return redisDatabase
}

func getRedisAddress() string {
	network := os.Getenv("REDIS_ADDR")
	port := os.Getenv("REDIS_PORT")
	return fmt.Sprintf("%s:%s", network, port)
}


func InitRedistestDb() redis.Cmdable {
	if redisTestDatabase != nil {
		return redisTestDatabase
	}

	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	redisTestDatabase =  redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return redisTestDatabase
}