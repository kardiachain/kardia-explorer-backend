// Package main
package main

import (
	"context"
	"go.uber.org/zap"
	"time"

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
				srv.Logger.Warn("Cannot pop unverified block height", zap.Error(err))
				continue
			}
			lgr := srv.Logger.With(zap.Uint64("block", blockHeight))
			lgr.Debug("Checking block integrity...")
			isValid, err := srv.VerifyBlock(ctx, blockHeight)
			if err != nil {
				lgr.Warn("Error while verifying block", zap.Error(err))
				continue
			}
			if !isValid {
				lgr.Warn("Block integrity is violated, reimporting...")
				err := srv.UpsertBlock(ctx, blockHeight)
				if err != nil {
					lgr.Warn("Block integrity is violated, reimporting...")
				}
			}
		}
	}
}
