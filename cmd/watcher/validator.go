package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
)

//syncValidators call each interval timer and fetch new validators from network
func watchValidators(ctx context.Context, w server.ValidatorWatcher, interval time.Duration) {
	lgr, _ := zap.NewProduction()
	lgr = lgr.With(zap.String("service", "watcher"))
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := w.SyncValidators(ctx); err != nil {
				lgr.Warn("cannot sync validator", zap.Error(err))
			}

		}
	}
}
