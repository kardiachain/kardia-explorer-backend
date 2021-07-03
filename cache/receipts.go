// Package cache
package cache

import (
	"context"
)

const (
	KeyPendingReceipts = "receipts#pending"
)

type IReceipts interface {
	PushReceipts(ctx context.Context, hashes []string) error
	PopReceipt(ctx context.Context) (string, error)
}

func (c *Redis) PushReceipts(ctx context.Context, hashes []string) error {
	var insertList []interface{}
	for _, h := range hashes {
		insertList = append(insertList, h)
	}
	_, err := c.client.LPush(ctx, KeyPendingReceipts, insertList...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PopReceipt(ctx context.Context) (string, error) {
	hash, err := c.client.LPop(ctx, KeyPendingReceipts).Result()
	if err != nil {
		return "", err
	}

	return hash, nil
}
