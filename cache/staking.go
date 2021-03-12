// Package cache
package cache

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	keyStakingStats = "#staking#stats"
)

type IStaking interface {
	UpdateStakingStats(ctx context.Context, stats *types.StakingStats) error
	StakingStats(ctx context.Context) (*types.StakingStats, error)
}

func (c *Redis) UpdateStakingStats(ctx context.Context, stats *types.StakingStats) error {
	lgr := c.logger.With(zap.String("method", "UpdateStakingStats"))
	lgr.Info("Update staking stats", zap.Any("stats", stats))
	data, err := json.Marshal(stats)
	if err != nil {
		lgr.Error("cannot marshal data", zap.Error(err))
		return err
	}
	if _, err := c.client.Set(ctx, keyStakingStats, string(data), 0).Result(); err != nil {
		lgr.Error("cannot set stats", zap.Error(err))
		return err
	}

	return nil
}

func (c *Redis) StakingStats(ctx context.Context) (*types.StakingStats, error) {
	result, err := c.client.Get(ctx, keyStakingStats).Result()
	if err != nil {
		return nil, err
	}
	var stats *types.StakingStats
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		return nil, err
	}
	// get current circulating supply that we updated manually, if exists
	return stats, nil
}
