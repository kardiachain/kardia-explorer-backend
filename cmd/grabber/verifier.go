// Package main
package main

import (
	"context"
	"go.uber.org/zap"
	"time"

	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/types"
)

func verify(ctx context.Context, srv *server.Server, interval time.Duration) {
	verifyFunc := func(db, network *types.Block) bool {
		if srv.VerifyBlockParam.VerifyTxCount {
			if db.NumTxs != network.NumTxs {
				return false
			}
		}
		if srv.VerifyBlockParam.VerifyBlockHash {
			return true
		}
		return true
	}
	srv.Logger.Info("Start verifying data...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			blockHeight, err := srv.InfoServer().PopUnverifiedBlockHeight(ctx)
			if err != nil {
				srv.Logger.Debug("Verifier: Cannot pop unverified block height from cache", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			lgr.Debug("Verifier: Checking block integrity...")
			networkBlock, err := srv.InfoServer().BlockByHeightFromRPC(ctx, blockHeight)
			if err != nil {
				lgr.Warn("Verifier: Error while get compare block from RPC, re-inserting this block to unverified list...", zap.Error(err))
				_ = srv.InfoServer().InsertUnverifiedBlocks(ctx, blockHeight)
				continue
			}
			result, err := srv.InfoServer().VerifyBlock(ctx, blockHeight, networkBlock, verifyFunc)
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
