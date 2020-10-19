// Package main
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/server"
)

func processReceipt(ctx context.Context, srv *server.Server) {
	logger := srv.Logger.With(zap.String("service", "receipt"))
	logger.Info("Start process receipts...")
	for {
		// Create worker base config

		// Pop tx hash from receipt queue and start upsert

		// If error then push bach

	}
}
