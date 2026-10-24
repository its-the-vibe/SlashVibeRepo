package redis

import (
	"github.com/redis/go-redis/v9"
)

func NewClient(address, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
}
