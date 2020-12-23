// Package main
package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
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
				srv.Logger.Debug("Verifier: Cannot pop unverified block height from cache", zap.Error(err))
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
