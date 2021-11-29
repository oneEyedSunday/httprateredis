package httprateredis

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/httprate"
	"github.com/go-redis/redis/v8"
)

type redisRateLimiter struct {
	inner        *redis.Client
	mu           sync.Mutex
	windowLength time.Duration
}

func NewRedisRateLimiter(c *redis.Client, windowLength time.Duration) *redisRateLimiter {
	return newRedisRateLimiter(c, windowLength)
}

func newRedisRateLimiter(c *redis.Client, windowLength time.Duration) *redisRateLimiter {
	return &redisRateLimiter{
		inner:        c,
		windowLength: windowLength,
	}
}

func (c *redisRateLimiter) Increment(key string, currentWindow time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	hkey := string(httprate.LimitCounterKey(key, currentWindow))

	c.inner.Incr(context.TODO(), hkey)
	c.inner.Expire(context.TODO(), hkey, c.windowLength)
	return nil
}

func (c *redisRateLimiter) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	curr, err := c.inner.Get(context.TODO(), string(httprate.LimitCounterKey(key, currentWindow))).Result()

	if err != nil {
		return 0, 0, err
	}

	prev, err := c.inner.Get(context.TODO(), string(httprate.LimitCounterKey(key, previousWindow))).Result()

	if err != nil {
		return 0, 0, err
	}

	currInt, err := strconv.Atoi(curr)
	if err != nil {
		return 0, 0, err
	}

	prevInt, err := strconv.Atoi(prev)
	if err != nil {
		return currInt, 0, err
	}

	return currInt, prevInt, nil
}

var _ httprate.LimitCounter = &redisRateLimiter{}
