// Package db
package db

import (
	"context"

	"github.com/kardiachain/go-kardia/types/time"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
)

const (
	cStats = "Stats"
)

func (m *mongoDB) UpdateDailyStats(ctx context.Context) {
	m.wrapper.C(cStats)
}

func (m *mongoDB) UpdateStats(ctx context.Context, stats *types.Stats) error {
	_, err := m.wrapper.C(cStats).Insert(stats)
	if err != nil {
		return err
	}
	// remove old stats
	if _, err := m.wrapper.C(cStats).RemoveAll(bson.M{"updatedAtBlock": bson.M{"$lt": stats.UpdatedAtBlock}}); err != nil {
		m.logger.Warn("cannot remove old stats", zap.Error(err), zap.Uint64("latest updated block", stats.UpdatedAtBlock))
		return err
	}
	return nil
}

func (m *mongoDB) Stats(ctx context.Context) *types.Stats {
	var stats *types.Stats
	if err := m.wrapper.C(cStats).FindOne(bson.M{}).Decode(&stats); err == nil {
		// remove blocks after checkpoint
		latestBlock, err := m.Blocks(ctx, &types.Pagination{
			Skip:  0,
			Limit: 1,
		})
		if len(latestBlock) > 0 {
			stats.UpdatedAtBlock = latestBlock[0].Height
		}
		for {
			if stats.UpdatedAtBlock%cfg.UpdateStatsInterval == 0 {
				break
			}
			stats.UpdatedAtBlock, err = m.DeleteLatestBlock(ctx)
			if err != nil {
				m.logger.Warn("Getting stats: DeleteLatestBlock error", zap.Error(err))
			}
			stats.UpdatedAtBlock--
		}
		return stats
	}
	// create a checkpoint (latestBlockHeight) and remove blocks after checkpoint
	// then calculate stats based on current database
	latestBlockHeight := uint64(0)
	latestBlock, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if len(latestBlock) > 0 {
		latestBlockHeight = latestBlock[0].Height
	}
	for {
		if latestBlockHeight%cfg.UpdateStatsInterval == 0 {
			break
		}
		latestBlockHeight, err = m.DeleteLatestBlock(ctx)
		if err != nil {
			m.logger.Warn("Getting stats: DeleteLatestBlock error", zap.Error(err))
		}
		latestBlockHeight--
	}
	totalAddrs, totalContracts, err := m.GetTotalAddresses(ctx)
	if err != nil {
		totalAddrs = 0
		totalContracts = 0
	}
	totalTxs, err := m.TxsCount(ctx)
	if err != nil {
		totalTxs = 0
	}
	return &types.Stats{
		UpdatedAt:         time.Now(),
		UpdatedAtBlock:    latestBlockHeight,
		TotalTransactions: totalTxs,
		TotalAddresses:    totalAddrs,
		TotalContracts:    totalContracts,
	}
}
