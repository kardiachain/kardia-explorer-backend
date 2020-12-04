// Package main
package main

import (
	"context"
	"time"

	"github.com/kardiachain/explorer-backend/server"
	"go.uber.org/zap"
)

var (
	currentProcessBlock uint64
	processCounter      = 0
	startTime           time.Time
	endTime             time.Duration
)

const (
	counterLimit = 3
)

func IsSkip() bool {
	return processCounter >= counterLimit
}

func backfill(ctx context.Context, srv *server.Server, interval time.Duration) {
	srv.Logger.Info("Start refilling...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.InfoServer().PopErrorBlockHeight(ctx)
			if blockHeight == currentProcessBlock && blockHeight != 0 {
				processCounter++
				if IsSkip() {
					srv.Logger.Warn("Refilling: Skip block since several error attempts, inserting to persistent error blocks list", zap.Uint64("BlockHeight", blockHeight))
					_ = srv.InfoServer().InsertPersistentErrorBlocks(ctx, blockHeight)
					// Reset counter
					processCounter = 0
					continue
				}
			}
			currentProcessBlock = blockHeight
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			if err != nil {
				lgr.Debug("Refilling: Failed to pop error block number", zap.Error(err))
				_ = srv.InfoServer().InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if blockHeight == 0 {
				continue
			}
			lgr.Info("Refilling:")
			// insert current block height to cache for re-verifying later
			err = srv.InfoServer().InsertUnverifiedBlocks(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to insert unverified block", zap.Error(err))
			}
			// try to get block
			block, err := srv.InfoServer().BlockByHeight(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to get block", zap.Error(err))
				_ = srv.InfoServer().InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			// upsert this block to database only
			startTime = time.Now()
			if err := srv.InfoServer().UpsertBlock(ctx, block); err != nil {
				lgr.Error("Refilling: Failed to upsert block", zap.Error(err))
				_ = srv.InfoServer().InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			endTime = time.Since(startTime)
			srv.Metrics().RecordUpsertBlockTime(endTime)
			lgr.Debug("Refilling: Upsert block time", zap.Duration("TimeConsumed", endTime), zap.String("Avg", srv.Metrics().GetUpsertBlockTime()))
		}
	}
}
