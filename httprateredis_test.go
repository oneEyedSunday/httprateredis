package httprateredis_test

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/oneeyedsunday/httprateredis"
	"github.com/stretchr/testify/assert"
)

func TestRedisLimit(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:8379",
	})
	r := httprateredis.NewRedisRateLimiter(rdb, time.Second*5)
	k := "bar"
	w := time.Now()

	t.Run("inits to zero", func(t *testing.T) {
		curr, prev, err := r.Get(k, time.Now().Add(time.Second), time.Now())
		assert.Equal(t, curr, 0)
		assert.Equal(t, prev, 0)
		assert.Equal(t, err, nil)
	})

	t.Run("increments", func(t *testing.T) {
		err := r.Increment(k, w.Add(time.Millisecond*700))
		assert.Equal(t, err, nil)
	})
}
