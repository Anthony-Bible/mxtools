package secondary

import (
	"context"
	"errors"
	"time"

	"mxclone/pkg/redisiface"

	"github.com/redis/go-redis/v9"
)

// Adapter to make *redis.Client satisfy internal.redisClient interface
// and to allow use in production InitJobStore

type RedisClientAdapter struct {
	Client *redis.Client
}

func (a *RedisClientAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.Client.Set(ctx, key, value, expiration).Err()
}

func (a *RedisClientAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.Client.Get(ctx, key).Result()
}

func (a *RedisClientAdapter) Watch(ctx context.Context, fn func(redisiface.RedisClient) error, keys ...string) error {
	return a.Client.Watch(ctx, func(tx *redis.Tx) error {
		return fn(&RedisTxAdapter{Tx: tx})
	}, keys...)
}

func (a *RedisClientAdapter) TxPipelined(ctx context.Context, fn func(redisiface.RedisClient) error) error {
	_, err := a.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		return fn(a) // Use the same adapter for pipeliner
	})
	return err
}

func (a *RedisClientAdapter) Close() error {
	return a.Client.Close()
}

// Adapter for redis.Tx to satisfy internal.redisClient interface for Watch
// Only implements Get, Set, TxPipelined, Close as needed
// (TxPipelined on Tx just calls the function directly)
type RedisTxAdapter struct {
	Tx *redis.Tx
}

func (a *RedisTxAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.Tx.Set(ctx, key, value, expiration).Err()
}
func (a *RedisTxAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.Tx.Get(ctx, key).Result()
}
func (a *RedisTxAdapter) Watch(ctx context.Context, fn func(redisiface.RedisClient) error, keys ...string) error {
	return errors.New("nested Watch not supported")
}
func (a *RedisTxAdapter) TxPipelined(ctx context.Context, fn func(redisiface.RedisClient) error) error {
	_, err := a.Tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		return fn(a) // Use same adapter
	})
	return err
}
func (a *RedisTxAdapter) Close() error { return nil }
