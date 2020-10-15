// Package cache
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Adapter string

const (
	RedisAdapter Adapter = "redis"
)

type Config struct {
	Adapter Adapter
	URL     string
	DB      int

	IsFlush bool

	BlockBuffer        int64
	DefaultExpiredTime time.Duration

	Logger *zap.Logger
}

type Client interface {
	InsertBlock(ctx context.Context, block *types.Block) error
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)

	InsertTxs(ctx context.Context, txs []*types.Transaction) error
	TxByHash(ctx context.Context, txHash string) (*types.Transaction, error)

	BlocksSize(ctx context.Context) (int64, error)
}

func New(cfg Config) (Client, error) {
	cfg.Logger.Debug("create cache client with config", zap.Any("config", cfg))
	switch cfg.Adapter {
	case RedisAdapter:
		return newRedis(cfg)
	}
	return nil, errors.New("invalid cache config")
}

func newRedis(cfg Config) (Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.URL,
		DB:   cfg.DB,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}
	if cfg.IsFlush {
		msg, err := redisClient.FlushAll(context.Background()).Result()
		if err != nil || msg != "OK" {
			return nil, err
		}
	}

	logger := cfg.Logger.With(zap.String("cache", "redis"))
	client := &Redis{
		client: redisClient,
		logger: logger,
	}
	client.cfg = cfg
	return client, nil
}
