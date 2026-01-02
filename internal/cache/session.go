package cache

import "github.com/go-redis/redis/v8"

// getSession returns a Redis pipeline session.
func (c *Cache) getSession() redis.Pipeliner {
	return c.RedisClient.Pipeline()
}
