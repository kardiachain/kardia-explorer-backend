// Package main
package main

import (
	"context"
	"errors"
	"time"

	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/types"

	"go.uber.org/zap"
)

var (
	currentProcessBlock uint64
	processCounter      = 0
)

const (
	counterLimit = 3
)

func IsSkip() bool {
	return processCounter >= counterLimit
}

func backfill(ctx context.Context, srv *server.Server) {
	srv.Logger.Info("Start refilling...")
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.PopErrorBlockHeight(ctx)
			if blockHeight == currentProcessBlock && blockHeight != 0 && processCounter != 0 {
				processCounter++
				if IsSkip() {
					srv.Logger.Warn("Skip block since expected error", zap.Uint64("BlockHeight", blockHeight))
					// Reset counter
					processCounter = 0
					continue
				}
			}
			currentProcessBlock = blockHeight
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			if err != nil {
				lgr.Info("Refilling: Failed to pop error block number", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if blockHeight == 0 {
				continue
			}
			lgr.Info("Refilling: ")
			block, err := srv.BlockByHeight(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to get block", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if block == nil {
				lgr.Error("Refilling: Block not found")
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if err := srv.ImportBlock(ctx, block, false); err != nil {
				if !errors.Is(err, types.ErrRecordExist) {
					lgr.Error("Record exist")
				}
				lgr.Error("Refilling: Failed to import block", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
		}
	}
}
