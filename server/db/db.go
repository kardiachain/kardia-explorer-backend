// Package db
package db

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Adapter string

const (
	MGO Adapter = "mongo"
	PG  Adapter = "postgres"
)

type ClientConfig struct {
	DbAdapter Adapter
	DbName    string
	URL       string
	FlushDB   bool

	Logger *zap.Logger
}

// DB define list API used by infoServer
type Client interface {
	ping() error
	InsertBlock(ctx context.Context, block *types.Block) error
	UpsertBlock(ctx context.Context, block *types.Block) error
	InsertTxs(ctx context.Context, txs []*types.Transaction) error
	UpsertTxs(ctx context.Context, txs []*types.Transaction) error
	UpdateActiveAddress() error
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	IsBlockExist(ctx context.Context, block *types.Block) (bool, error)
}

func NewClient(cfg ClientConfig) (Client, error) {
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid config")
	}
}
