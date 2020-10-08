// Package cache
package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Redis struct {
	client *redis.Client

	logger *zap.Logger
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber return blockIndex
func (c *Redis) ImportBlock(ctx context.Context, block *types.Block) error {
	// Push new block to list
	lPushResult := c.client.LPush(ctx, KeyBlocks, block.String())
	blockIndex, err := lPushResult.Result()
	if err != nil {
		return err
	}

	c.logger.Debug("Push new block success", zap.Int64("index", blockIndex))
	blockByHashKey := fmt.Sprintf(KeyBlockByHash, block.Hash)
	blockByNumberKey := fmt.Sprintf(KeyBlockByNumber, block.Height)
	if err := c.client.Set(ctx, blockByHashKey, blockIndex, 0).Err(); err != nil {
		return err
	}
	if err := c.client.Set(ctx, blockByNumberKey, blockIndex, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	key := fmt.Sprintf(KeyBlockByHash, blockHash)
	var index int64
	if err := c.client.Get(ctx, key).Scan(&index); err != nil {
		return nil, err
	}
	return c.getBlockIndex(ctx, index)
}

func (c *Redis) BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	key := fmt.Sprintf(KeyBlockByNumber, blockHeight)
	var index int64
	if err := c.client.Get(ctx, key).Scan(&index); err != nil {
		return nil, err
	}
	return c.getBlockIndex(ctx, index)
}

func (c *Redis) getBlockIndex(ctx context.Context, index int64) (*types.Block, error) {
	block := &types.Block{}
	lIndexResult := c.client.LIndex(ctx, KeyBlocks, index)
	if err := lIndexResult.Err(); err != nil {
		return nil, err
	}

	if err := lIndexResult.Scan(block); err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *Redis) LatestBlock() {

}
