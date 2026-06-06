package cache

import "github.com/redis/go-redis/v9"

// GetSession returns a Redis pipeline session.
func (c *Cache) GetSession() redis.Pipeliner {
	return c.RedisClient.Pipeline()
}
