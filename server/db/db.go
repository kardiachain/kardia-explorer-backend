// Package db
package db

import (
	"context"

	"github.com/kardiachain/explorer-backend/types"
)

// DB define list API used by infoServer
type Client interface {
	ping() error
	ImportBlock(ctx context.Context, block *types.Block) error
	UpdateActiveAddress() error
	BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error)
}
