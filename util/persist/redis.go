package persist

import (
	"time"

	"github.com/go-redis/redis"
)

func NewRedisClient(addr string, pwd string) *redis.Client {
	const maxPoolSize = 80

	c := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pwd,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     maxPoolSize,
		PoolTimeout:  31 * time.Second,
	})
	_, err := c.Ping().Result()
	if err != nil {
		panic(err)
	}
	return c
}
