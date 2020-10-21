// Package main
package main

import (
	"context"
	"time"

	"github.com/kardiachain/explorer-backend/server"
	"go.uber.org/zap"
)

func backfill(ctx context.Context, srv *server.Server) {
	srv.Logger.Info("Start refilling...")
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.PopErrorBlockHeight(ctx)
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			lgr.Debug("Refilling: blocks")
			if err != nil {
				lgr.Error("Refilling: Failed to pop error block number", zap.Error(err))
				go srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			// TODO(trinhdn): remove hardcode
			if blockHeight == 0 {
				continue
			}
			block, err := srv.BlockByHeight(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to get block", zap.Error(err))
				go srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if block == nil {
				lgr.Error("Refilling: Block not found")
				go srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if err := srv.ImportBlock(ctx, block, false); err != nil {
				lgr.Error("Refilling: Failed to import block", zap.Error(err))
				go srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
		}
	}
}
