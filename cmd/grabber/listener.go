// Package main
package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
)

// listener fetch LatestBlockNumber every second and check if we stay behind latest block
// todo: implement pipeline with worker for dispatch InsertBlock task
func listener(ctx context.Context, srv *server.Server) {
	srv.Logger.Info("Start listening...")
	var (
		prevHeader uint64 = 0
		startTime  time.Time
		endTime    time.Duration
	)
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			latest, err := srv.LatestBlockHeight(ctx)
			srv.Logger.Debug("Get block height from network", zap.Uint64("BlockHeight", latest), zap.Uint64("PrevHeader", prevHeader))
			if err != nil {
				srv.Logger.Error("Listener: Failed to get latest block number", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", latest))

			// todo @longnd: this check quite bad, since its require us to keep backfill running
			// for example, if our
			if prevHeader != latest {
				startTime = time.Now()
				block, err := srv.BlockByHeight(ctx, latest)
				if err != nil {
					lgr.Error("Listener: Failed to get block", zap.Error(err))
					lgr.Debug("Block not found result", zap.Error(err))
					continue
				}
				endTime = time.Since(startTime)
				srv.Metrics().RecordScrapingTime(endTime)
				lgr.Info("Listener: Scraping block time", zap.Duration("TimeConsumed", endTime), zap.String("Avg", srv.Metrics().GetScrapingTime()))
				if block == nil {
					lgr.Error("Listener: Block not found")
					continue
				}
				if err := srv.ImportBlock(ctx, block, true); err != nil {
					lgr.Error("Listener: Failed to import block", zap.Error(err))
					continue
				}
				if latest-1 > prevHeader {
					lgr.Warn("Listener: We are behind network, inserting error blocks", zap.Uint64("from", prevHeader), zap.Uint64("to", latest))
					err := srv.InsertErrorBlocks(ctx, prevHeader, latest)
					if err != nil {
						lgr.Error("Listener: Failed to insert error block height", zap.Error(err))
						continue
					}
				}
				prevHeader = latest
			}
		}
	}
}
