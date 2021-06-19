// Package cache
package cache

import (
	"context"

	"github.com/kardiachain/go-kardia/lib/common"
)

const (
	KeyAddressSyncList = "#addresses#sync"
)

type IAddress interface {
	PushAddress(ctx context.Context, addresses []interface{}) error
	PopAddress(ctx context.Context) (string, error)
}

//PushAddress
func (c *Redis) PushAddress(ctx context.Context, addresses []interface{}) error {
	if len(addresses) == 0 {
		return nil
	}
	_, err := c.client.LPush(ctx, KeyAddressSyncList, addresses...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PopAddress(ctx context.Context) (string, error) {
	address, err := c.client.LPop(ctx, KeyAddressSyncList).Result()
	if err != nil {
		return "", err
	}
	// Make sure address is checksum
	return common.HexToAddress(address).String(), nil
}
