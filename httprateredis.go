package httprateredis

import (
	"context"
	"fmt"
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

	hkey := uintToString(httprate.LimitCounterKey(key, currentWindow))
	ctx, cancelFunc := c.getContextWithTimeout()
	defer cancelFunc()

	pipe := c.inner.Pipeline()
	pipe.Expire(ctx, hkey, c.windowLength*3)
	pipe.Incr(ctx, hkey)
	_, err := pipe.Exec(ctx)

	return err
}

func (c *redisRateLimiter) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancelFunc := c.getContextWithTimeout()
	defer cancelFunc()

	pipe := c.inner.Pipeline()
	pipe.Get(ctx, uintToString(httprate.LimitCounterKey(key, currentWindow)))
	pipe.Get(ctx, uintToString(httprate.LimitCounterKey(key, previousWindow)))
	res, err := pipe.Exec(ctx)
	if err != nil {
		curr, _ := strconv.Atoi(res[0].String())
		prev, _ := strconv.Atoi(res[1].String())

		return curr, prev, nil
	}

	return 0, 0, err
}

func (c *redisRateLimiter) getContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), c.timeout)
}

func uintToString(val uint64) string {
	return fmt.Sprintf("%v", val)
}

var _ httprate.LimitCounter = &redisRateLimiter{}
