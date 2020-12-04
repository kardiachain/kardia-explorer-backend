// Package main
package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
)

var prevHeader uint64 = 0 // the highest persistent block in database, don't need to backfill blocks have blockHeight < prevHeader

// listener fetch LatestBlockNumber every second and check if we stay behind latest block
// todo: implement pipeline with worker for dispatch InsertBlock task
func listener(ctx context.Context, srv *server.Server, interval time.Duration) {
	var (
		startTime time.Time
		endTime   time.Duration
	)
	// delete current latest block in db
	deletedHeight, err := srv.InfoServer().DeleteLatestBlock(ctx)
	if err != nil {
		srv.Logger.Warn("Cannot remove old latest block", zap.Error(err))
	}
	if deletedHeight > 0 {
		prevHeader = deletedHeight - 1 // the highest persistent block in database now is deletedHeight - 1
	}
	srv.Logger.Info("Start listening...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			latest, err := srv.InfoServer().LatestBlockHeight(ctx)
			srv.Logger.Debug("Listener: Get block height from network", zap.Uint64("BlockHeight", latest), zap.Uint64("PrevHeader", prevHeader))
			if err != nil {
				srv.Logger.Error("Listener: Failed to get latest block number", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", latest))
			if latest <= prevHeader {
				srv.Logger.Debug("Listener: No new block from RPC", zap.Uint64("prevHeader", prevHeader))
				continue
			}
			// todo @longnd: this check quite bad, since its require us to keep backfill running
			if prevHeader != latest {
				startTime = time.Now()
				block, err := srv.InfoServer().BlockByHeight(ctx, latest)
				if err != nil {
					lgr.Error("Listener: Failed to get block from RPC", zap.Error(err))
					continue
				}
				endTime = time.Since(startTime)
				srv.Metrics().RecordScrapingTime(endTime)
				lgr.Info("Listener: Scraping block time", zap.Duration("TimeConsumed", endTime), zap.String("Avg", srv.Metrics().GetScrapingTime()))
				if block == nil {
					lgr.Error("Listener: Block not found")
					continue
				}
				// insert current block height to cache for re-verifying later
				err = srv.InfoServer().InsertUnverifiedBlocks(ctx, latest)
				if err != nil {
					lgr.Error("Listener: Failed to insert unverified block", zap.Error(err))
				}
				// import this latest block to cache and database
				if err := srv.InfoServer().ImportBlock(ctx, block, true); err != nil {
					lgr.Error("Listener: Failed to import block", zap.Error(err))
					continue
				}
				if latest-1 > prevHeader {
					lgr.Warn("Listener: We are behind network, inserting error blocks", zap.Uint64("from", prevHeader), zap.Uint64("to", latest))
					err := srv.InfoServer().InsertErrorBlocks(ctx, prevHeader, latest)
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
