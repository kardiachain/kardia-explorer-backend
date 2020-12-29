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

	"github.com/kardiachain/explorer-backend/server"
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

func backfill(ctx context.Context, srv *server.Server, interval time.Duration) {
	srv.Logger.Info("Start refilling...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.PopErrorBlockHeight(ctx)
			if blockHeight == currentProcessBlock && blockHeight != 0 {
				processCounter++
				if IsSkip() {
					srv.Logger.Warn("Refilling: Skip block since several error attempts, inserting to persistent error blocks list", zap.Uint64("BlockHeight", blockHeight))
					_ = srv.InsertPersistentErrorBlocks(ctx, blockHeight)
					// Reset counter
					processCounter = 0
					continue
				}
			}
			currentProcessBlock = blockHeight
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			if err != nil {
				lgr.Debug("Refilling: Failed to pop error block number", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			if blockHeight == 0 {
				continue
			}
			lgr.Info("Refilling:")
			// insert current block height to cache for re-verifying later
			err = srv.InsertUnverifiedBlocks(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to insert unverified block", zap.Error(err))
			}
			// try to get block
			block, err := srv.BlockByHeight(ctx, blockHeight)
			if err != nil {
				lgr.Error("Refilling: Failed to get block", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
			// upsert this block to database only
			if err := srv.UpsertBlock(ctx, block); err != nil {
				lgr.Error("Refilling: Failed to upsert block", zap.Error(err))
				_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
				continue
			}
		}
	}
}
