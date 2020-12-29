/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
// Package main
package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/server"
)

var prevHeader uint64 = 0 // the highest persistent block in database, don't need to backfill blocks have blockHeight < prevHeader

// listener fetch LatestBlockNumber every second and check if we stay behind latest block
func listener(ctx context.Context, srv *server.Server, interval time.Duration) {
	var (
		startTime time.Time
		endTime   time.Duration
	)
	// update current stats of network and get highest persistent block in database
	prevHeader = srv.GetCurrentStats(ctx)
	srv.Logger.Info("Start listening...", zap.Uint64("from block", prevHeader))
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			latest, err := srv.LatestBlockHeight(ctx)
			srv.Logger.Info("Listener: Get block height from network", zap.Uint64("BlockHeight", latest), zap.Uint64("PrevHeader", prevHeader))
			if err != nil {
				srv.Logger.Error("Listener: Failed to get latest block number", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", latest))
			if latest <= prevHeader {
				continue
			}
			if prevHeader != latest {
				startTime = time.Now()
				block, err := srv.BlockByHeight(ctx, latest)
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
				err = srv.InsertUnverifiedBlocks(ctx, latest)
				if err != nil {
					lgr.Error("Listener: Failed to insert unverified block", zap.Error(err))
				}
				// import this latest block to cache and database
				if err := srv.ImportBlock(ctx, block, true); err != nil {
					lgr.Debug("Listener: Failed to import block", zap.Error(err))
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
				if latest%cfg.UpdateStatsInterval == 0 {
					_ = srv.UpdateCurrentStats(ctx)
				}
			}
		}
	}
}
