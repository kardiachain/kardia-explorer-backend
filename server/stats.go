// Package server
package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cache"
	"github.com/kardiachain/explorer-backend/db"
	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/types"
)

type StatsWatcher interface {
	BlockByHeight(ctx context.Context, height uint64) (*types.Block, error)
	UpdateStats(ctx context.Context, stats *types.DailyStats) error
}

type statsWatcher struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	lgr *zap.Logger
}

func NewStatsServer() (StatsWatcher, error) {
	return &statsWatcher{}, nil
}

func (s *statsWatcher) BlockByHeight(ctx context.Context, height uint64) (*types.Block, error) {
	block, err := s.dbClient.BlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (s *statsWatcher) CalculateTxsStats(ctx context.Context, block *types.Block) {

}

func (s *statsWatcher) UpdateStats(ctx context.Context, stats *types.DailyStats) error {
	return nil
}
