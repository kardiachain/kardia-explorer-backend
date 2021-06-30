// Package cache
package cache

import (
	"context"
)

type IReceipts interface {
	PushReceipts(ctx context.Context, hashes []string) error
	PopReceipt(ctx context.Context) (string, error)
}

func (c *Redis) PushReceipts(ctx context.Context, hashes []string) error {
	return nil
}

func (c *Redis) PopReceipt(ctx context.Context) (string, error) {
	return "", nil
}
