package cache

import "github.com/redis/go-redis/v9"

// getSession returns a Redis pipeline session.
func (c *Cache) getSession() redis.Pipeliner {
	return c.RedisClient.Pipeline()
}
