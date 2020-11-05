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

func backfill(ctx context.Context, srv *server.Server) {
	srv.Logger.Info("Start refilling...")
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.PopErrorBlockHeight(ctx)
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			if err != nil {
				lgr.Info("Refilling: Failed to pop error block number", zap.Error(err))
				err := srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				if err != nil {
					lgr.Error("Listener: Failed to insert error block height", zap.Error(err))
					continue
				}
			}
			lgr.Info("Refilling: ")
			// TODO(trinhdn): remove hardcode
			if blockHeight == 0 {
				continue
			}
			block, err := srv.BlockByHeight(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to get block", zap.Error(err))
				err := srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				if err != nil {
					lgr.Error("Listener: Failed to insert error block height", zap.Error(err))
					continue
				}
			}
			if block == nil {
				lgr.Error("Refilling: Block not found")
				err := srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				if err != nil {
					lgr.Error("Listener: Failed to insert error block height", zap.Error(err))
					continue
				}
			}
			if err := srv.ImportBlock(ctx, block, false); err != nil {
				if !errors.Is(err, types.ErrRecordExist) {
					continue
				}
				lgr.Error("Refilling: Failed to import block", zap.Error(err))
				err := srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				if err != nil {
					lgr.Error("Listener: Failed to insert error block height", zap.Error(err))
					continue
				}

			}
		}
	}
}
