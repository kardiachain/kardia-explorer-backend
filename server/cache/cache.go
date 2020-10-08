// Package cache
package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Adapter string

const (
	RedisAdapter Adapter = "redis"
)

type Config struct {
	RedisUrl string

	Logger *zap.Logger
}

const (
	KeyBlocks = "#blocks" // List

	KeyLatestStats = "#stats#latest"

	KeyLatestBlockNumber = "#block#latestNumber"
	KeyLatestBlock       = "#block#latest"
	KeyBlockByNumber     = "#block#%d"
	KeyBlockByHash       = "#block#%s"
)

type Client interface {
	ImportBlock(ctx context.Context, block *types.Block) error
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
}

func New(cfg Config) Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisUrl,
	})
	logger := cfg.Logger.With(zap.String("cache", "redis"))
	return &Redis{
		client: redisClient,
		logger: logger,
	}
}
