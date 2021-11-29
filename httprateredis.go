package httprateredis

import (
	"sync"
	"time"

	"github.com/go-chi/httprate"
	"github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
	inner     *redis.Client
	mu           sync.Mutex
}

func NewRedisRateLimiter(c *redis.Client) *RedisRateLimiter {
	return newRedisRateLimiter(c)
}

func newRedisRateLimiter(c *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		inner: c,
	}
}

func (c *RedisRateLimiter) Increment(key string, currentWindow time.Time) error {
	c.evict()
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil
}

func (c *RedisRateLimiter) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return 0,0, nil
}

func (c *RedisRateLimiter) evict() {
	c.mu.Lock()

	defer c.mu.Unlock()
}


var _ httprate.LimitCounter = &RedisRateLimiter{}
