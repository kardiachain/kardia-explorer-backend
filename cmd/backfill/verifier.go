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

	"github.com/kardiachain/kardia-explorer-backend/server"
)

func verify(ctx context.Context, srv *server.Server, interval time.Duration) {
	srv.Logger.Info("Start verifying data...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.PopUnverifiedBlockHeight(ctx)
			if err != nil {
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			lgr.Info("Verifier: Checking block integrity...")
			// get block by height from RPC in order to verify database block at the same height
			networkBlock, err := srv.BlockByHeightFromRPC(ctx, blockHeight)
			if err != nil {
				lgr.Warn("Verifier: Error while get compare block from RPC, re-inserting this block to unverified list...", zap.Error(err))
				_ = srv.InsertUnverifiedBlocks(ctx, blockHeight)
				continue
			}
			result, err := srv.VerifyBlock(ctx, blockHeight, networkBlock)
			if err != nil {
				lgr.Warn("Verifier: Error while verifying block", zap.Error(err))
				continue
			}
			if result {
				lgr.Warn("Verifier: Block in database is corrupted and successfully replaced")
			}
		}
	}
}
