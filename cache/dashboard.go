// Package cache
package cache

import (
	"context"
	"strconv"

	"github.com/kardiachain/kardia-explorer-backend/utils"
	"go.uber.org/zap"
)

const (
	KeyTotalHolders   = "#holders#total"
	KeyTotalContracts = "#contracts#total"
	KeyTotalAddresses = "#addresses#total"
)

type IDashboard interface {
	// KRC20Holders summary
	UpdateTotalHolders(ctx context.Context, holders uint64, contracts uint64) error
	TotalHolders(ctx context.Context) (uint64, uint64)

	UpdateTotalContracts(ctx context.Context, contracts int64) error
	TotalContracts(ctx context.Context) (int64, error)
	UpdateTotalAddresses(ctx context.Context, addresses int64) error
	TotalAddresses(ctx context.Context) (int64, error)
}

// KRC20Holders summary
func (c *Redis) UpdateTotalHolders(ctx context.Context, holders uint64, contracts uint64) error {
	if err := c.client.Set(ctx, KeyTotalHolders, holders, 0).Err(); err != nil {
		// Handle error here
		c.logger.Warn("cannot set total holders values")
	}
	if err := c.client.Set(ctx, KeyTotalContracts, contracts, 0).Err(); err != nil {
		// Handle error here
		c.logger.Warn("cannot set total contracts values")
	}
	return nil
}

func (c *Redis) TotalHolders(ctx context.Context) (uint64, uint64) {
	result, err := c.client.Get(ctx, KeyTotalHolders).Result()
	if err != nil {
		// Handle error here
	}
	// Convert to int
	totalHolders := utils.StrToUint64(result)

	result, err = c.client.Get(ctx, KeyTotalContracts).Result()
	if err != nil {
		// Handle error here
	}
	// Convert to int
	totalContracts := utils.StrToUint64(result)

	return totalHolders, totalContracts
}

func (c *Redis) UpdateTotalContracts(ctx context.Context, contracts int64) error {
	if err := c.client.Set(ctx, KeyTotalContracts, contracts, 0).Err(); err != nil {
		// Handle error here
		c.logger.Warn("cannot set total contracts values")
	}
	return nil
}

func (c *Redis) TotalContracts(ctx context.Context) (int64, error) {
	result, err := c.client.Get(ctx, KeyTotalContracts).Result()
	if err != nil {
		return 0, err
	}
	// Convert to int
	totalContracts, _ := strconv.ParseInt(result, 10, 64)
	return totalContracts, nil

}

func (c *Redis) UpdateTotalAddresses(ctx context.Context, addresses int64) error {
	if err := c.client.Set(ctx, KeyTotalAddresses, addresses, 0).Err(); err != nil {
		c.logger.Error("cannot set total addresses", zap.Error(err))
		return err
	}
	return nil
}

func (c *Redis) TotalAddresses(ctx context.Context) (int64, error) {
	result, err := c.client.Get(ctx, KeyTotalAddresses).Result()
	if err != nil {
		return 0, err
	}
	// Convert to int
	totalContracts, _ := strconv.ParseInt(result, 10, 64)
	return totalContracts, nil
}
