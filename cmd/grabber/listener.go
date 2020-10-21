// Package main
package main

import (
	"context"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
)

// listener fetch LatestBlockNumber every second and check if we stay behind latest block
// todo: implement pipeline with worker for dispatch InsertBlock task
func listener(ctx context.Context, srv *server.Server) {
	srv.Logger.Info("Start listening...")
	var prevHeader uint64 = 0
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			latest, err := srv.LatestBlockHeight(ctx)
			if err != nil {
				srv.Logger.Error("Listener: Failed to get latest block number", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", latest))
			// todo @longnd: this check quite bad, since its require us to keep backfill running
			// for example, if our
			if prevHeader != latest {
				lgr.Info("Listener: Getting block " + strconv.FormatUint(latest, 10))
				block, err := srv.BlockByHeight(ctx, latest)
				if err != nil {
					lgr.Error("Listener: Failed to get block", zap.Error(err))
					lgr.Debug("Block not found result", zap.Error(err))
					continue
				}
				if block == nil {
					lgr.Error("Listener: Block not found")
					continue
				}
				if err := srv.ImportBlock(ctx, block); err != nil {
					lgr.Error("Listener: Failed to import block", zap.Error(err))
					continue
				}
				if latest-prevHeader > 1 {
					lgr.Info("Listener: Insert error blocks", zap.Uint64("from", prevHeader), zap.Uint64("to", latest))
					go srv.InsertErrorBlocks(ctx, prevHeader, latest)
					continue
				}
				prevHeader = latest
			}
		}
	}
}
