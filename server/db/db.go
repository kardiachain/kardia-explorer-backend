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

	Logger *zap.Logger
}

// DB define list API used by infoServer
type Client interface {
	ping() error
	ImportBlock(ctx context.Context, block *types.Block) error
	UpdateActiveAddress() error
	BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error)
}

func NewClient(cfg ClientConfig) (Client, error) {
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid config")
	}
}
