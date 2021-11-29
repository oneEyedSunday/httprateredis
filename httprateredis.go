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
	timeout      time.Duration
}

func NewRedisRateLimiter(c *redis.Client, windowLength time.Duration) *redisRateLimiter {
	return newRedisRateLimiter(c, windowLength, time.Millisecond*50)
}

func NewRedisRateLimiterWithRedisTimeout(c *redis.Client, windowLength time.Duration, redisTimeout time.Duration) *redisRateLimiter {
	return newRedisRateLimiter(c, windowLength, redisTimeout)
}

func newRedisRateLimiter(c *redis.Client, windowLength time.Duration, redisTimeout time.Duration) *redisRateLimiter {
	return &redisRateLimiter{
		inner:        c,
		windowLength: windowLength,
		timeout:      redisTimeout,
	}
}

func (c *redisRateLimiter) Increment(key string, currentWindow time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	hkey := string(httprate.LimitCounterKey(key, currentWindow))
	ctx, cancelFunc := c.getContextWithTimeout()
	defer cancelFunc()

	c.inner.Incr(ctx, hkey)
	c.inner.Expire(ctx, hkey, c.windowLength)
	return nil
}

func (c *redisRateLimiter) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancelFunc := c.getContextWithTimeout()
	defer cancelFunc()

	curr, err := c.inner.Get(ctx, string(httprate.LimitCounterKey(key, currentWindow))).Result()

	if err != nil {
		return 0, 0, err
	}

	prev, err := c.inner.Get(ctx, string(httprate.LimitCounterKey(key, previousWindow))).Result()

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

func (c *redisRateLimiter) getContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), c.timeout)
}

var _ httprate.LimitCounter = &redisRateLimiter{}
