// Package main
package main

import (
	"context"
	"time"

	"github.com/kardiachain/kardia-explorer-backend/server"
	"go.uber.org/zap"
)

func balancer(ctx context.Context, srv *server.Server, interval time.Duration) {
	lgr := srv.Logger.With(zap.String("task", "balancer"))
	lgr.Info("Start balancer")
	t := time.NewTicker(interval)
	defer t.Stop()

}