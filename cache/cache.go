// Package cache
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
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
	IStaking
	IBlock
	ITxs

	// GetListHolders summary
	UpdateTotalHolders(ctx context.Context, holders uint64, contracts uint64) error
	TotalHolders(ctx context.Context) (uint64, uint64)

	IsRequestToCoinMarket(ctx context.Context) bool
	TokenInfo(ctx context.Context) (*types.TokenInfo, error)
	UpdateTokenInfo(ctx context.Context, tokenInfo *types.TokenInfo) error
	UpdateSupplyAmounts(ctx context.Context, supplyInfo *types.SupplyInfo) error

	Validators(ctx context.Context) (*types.Validators, error)
	UpdateValidators(ctx context.Context, validators *types.Validators) error

	SMCAbi(ctx context.Context, key string) (string, error)
	UpdateSMCAbi(ctx context.Context, key, abi string) error

	KRCTokenInfo(ctx context.Context, krcTokenAddr string) (*types.KRCTokenInfo, error)
	UpdateKRCTokenInfo(ctx context.Context, krcTokenInfo *types.KRCTokenInfo) error

	AddressInfo(ctx context.Context, addr string) (*types.Address, error)
	UpdateAddressInfo(ctx context.Context, addrInfo *types.Address) error

	ServerStatus(ctx context.Context) (*types.ServerStatus, error)
	UpdateServerStatus(ctx context.Context, serverStatus *types.ServerStatus) error

	CountBlocksOfProposer(ctx context.Context, proposerAddr string) (int64, error)
	UpdateNumOfBlocksByProposer(ctx context.Context, proposerAddr string, numOfBlocks int64) error
}

func New(cfg Config) (Client, error) {
	switch cfg.Adapter {
	case RedisAdapter:
		return newRedis(cfg)
	}
	return nil, errors.New("invalid cache config")
}

func newRedis(cfg Config) (Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.URL,
		DB:       cfg.DB,
		Password: "fengari@kaitothemoon123",
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
