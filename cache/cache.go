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
	IReceipts

	InsertBlock(ctx context.Context, block *types.Block) error
	InsertTxsOfBlock(ctx context.Context, block *types.Block) error
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHeight(ctx context.Context, blockHeight uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error)

	ListSize(ctx context.Context, key string) (int64, error)

	LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)
	LatestTransactions(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)

	InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error
	PopErrorBlockHeight(ctx context.Context) (uint64, error)
	InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error
	PersistentErrorBlockHeights(ctx context.Context) ([]uint64, error)
	InsertUnverifiedBlocks(ctx context.Context, height uint64) error
	PopUnverifiedBlockHeight(ctx context.Context) (uint64, error)

	UpdateTotalTxs(ctx context.Context, blockTxs uint64) (uint64, error)
	SetTotalTxs(ctx context.Context, numTxs uint64) error
	TotalTxs(ctx context.Context) uint64
	LatestBlockHeight(ctx context.Context) uint64

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
