// Package main
package main

import (
	"context"
	"time"

	"github.com/kardiachain/explorer-backend/server"
)

var UpdateAddressDefaultInterval = 3 * time.Minute

func updateAddresses(ctx context.Context, isUpdateContracts bool, blockRange uint64, srv *server.Server) {
	//lastUpdated := time.Now().Unix()
	t := time.NewTicker(UpdateAddressDefaultInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:

		}
	}
}
