package redisiface

import (
	"context"
	"time"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Watch(ctx context.Context, fn func(RedisClient) error, keys ...string) error
	TxPipelined(ctx context.Context, fn func(RedisClient) error) error
	Close() error
}
