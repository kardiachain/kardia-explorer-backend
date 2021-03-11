// Package handler
package handler

import (
	"context"
)

type IBlock interface {
	SubscribeNewBlock(ctx context.Context) error
}

func (h *handler) SubscribeNewBlock(ctx context.Context) error {
	return nil
}
