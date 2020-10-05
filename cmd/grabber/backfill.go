// Package main
package main

import (
	"context"
	"time"

	"github.com/kardiachain/network-explorer/server/utils"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/types"
)

func backfill(ctx context.Context, srv server.Server, blockNumber uint64) {
	var minBlockNum uint64 = 0
	var err error

	var validateBlockStrategy = func(db, network *types.Block) bool {
		return db.BlockHash != network.BlockHash
	}

	for {
		logger := srv.Logger.With(zap.Uint64("block", blockNumber))
		logger.Info("Backfilling...")
		if blockNumber == minBlockNum {
			blockNumber, err = srv.LatestBlockNumber(ctx)
			if err != nil {
				srv.Logger.Debug("cannot get latest block height", zap.Error(err))
				continue
			}
		}
		block := &types.Block{Height: blockNumber}
		if err := srv.ValidateBlock(ctx, block, validateBlockStrategy); err != nil {
			logger.Error("failed to validate block", zap.Error(err))
			return
		}
		if err != nil {
			logger.Error("Backfill: Failed to get block", zap.Error(err))
			if utils.SleepCtx(ctx, 5*time.Second) != nil {
				return
			}
			continue
		}
		blockNumber--
	}
}
