// Package db
package db

import (
	"context"

	"github.com/kardiachain/explorer-backend/types"
)

type MongoDB struct {
}

func (m *MongoDB) ping() error {
	panic("implement me")
}

func (m *MongoDB) importBlock(ctx context.Context, block *types.Block) error {
	panic("implement me")
}

func (m *MongoDB) updateActiveAddress() error {
	panic("implement me")
}
