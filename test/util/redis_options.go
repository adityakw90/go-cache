package testutil

type RedisTestOption func(*redisTestOptions)

type redisTestOptions struct {
	db int
}

func WithRedisDB(db int) RedisTestOption {
	return func(o *redisTestOptions) {
		o.db = db
	}
}
