// Package cache
package cache

import (
	"context"
)

type ITxs interface {
	UpdateTotalTxs(ctx context.Context, blockTxs uint64) (uint64, error)
	SetTotalTxs(ctx context.Context, numTxs uint64) error
	TotalTxs(ctx context.Context) uint64
	LatestBlockHeight(ctx context.Context) uint64
}
