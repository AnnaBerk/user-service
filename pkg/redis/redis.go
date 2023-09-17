package redis

import "github.com/go-redis/redis/v8"

func New(addr string, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	return rdb
}
