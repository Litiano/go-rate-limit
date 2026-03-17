package rate_limiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	redisClient      *redis.Client
	ctx              context.Context
	defaultRateLimit int64
	banDuration      time.Duration
}

func NewRedisRateLimiter(redisClient *redis.Client, defaultRateLimit int64, banDuration time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{redisClient: redisClient, ctx: context.Background(), defaultRateLimit: defaultRateLimit, banDuration: banDuration}
}

func (rl *RedisRateLimiter) GetRateLimitFor(subject string) (int64, error) {
	rateLimit, err := rl.redisClient.Get(rl.ctx, fmt.Sprintf("rate-limit-for:%v", subject)).Int64()
	if errors.Is(err, redis.Nil) {
		return rl.defaultRateLimit, nil
	}

	if err != nil {
		return 0, err
	}

	return rateLimit, nil
}

func (rl *RedisRateLimiter) GetRateCountFor(subject string) (int64, error) {
	// The Lua script for atomic increment and expiration
	luaScript := `
		local key = KEYS[1]
		local count = redis.call('INCR', key)
		if count == 1 then
			redis.call('EXPIRE', key, ARGV[1])
		end
		return count
	`
	key := fmt.Sprintf("rate-limit-count-for:%v", subject)
	expiration := 1 * time.Second

	count, err := rl.redisClient.Eval(rl.ctx, luaScript, []string{key}, expiration.Seconds()).Int64()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (rl *RedisRateLimiter) IsBanned(subject string) (bool, error) {
	result, err := rl.redisClient.Get(rl.ctx, fmt.Sprintf("rate-limit-banned:%v", subject)).Bool()

	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return result, nil
}

func (rl *RedisRateLimiter) Ban(subject string) error {
	_, err := rl.redisClient.Set(rl.ctx, fmt.Sprintf("rate-limit-banned:%v", subject), true, rl.banDuration).Result()

	return err
}

func (rl *RedisRateLimiter) Unban(subject string) error {
	_, err := rl.redisClient.Del(rl.ctx, fmt.Sprintf("rate-limit-banned:%v", subject)).Result()

	return err
}
