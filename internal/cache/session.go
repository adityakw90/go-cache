package cache

import "github.com/go-redis/redis/v8"

// GetSession returns a Redis pipeline session.
func (c *Cache) GetSession() redis.Pipeliner {
	return c.RedisClient.Pipeline()
}
