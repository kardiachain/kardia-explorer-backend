// Package cache
package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	KeyLatestBlockHeight = "#block#latestHeight"
	KeyLatestBlock       = "#block#latest"

	KeyBlocks       = "#blocks" // List
	KeyBlockKeyList = "#block#%s#keys"
	// Maintain a list to remove all key
	KeyBlockByNumber = "#block#%d"
	KeyBlockByHash   = "#block#%s"

	KeyTxsOfBlockIndex        = "#block#index#%d#txs"
	KeyTxsOfBlockIndexKeyList = "#block#index#%d#txs#keys"

	KeyTxOfBlockIndexByNonce = "#block#index#%d#tx#%d"
	KeyTxOfBlockIndexByHash  = "#block#index#%d#tx#%s"

	KeyLatestStats = "#stats#latest"
)

type Redis struct {
	client *redis.Client
	logger *zap.Logger
}

func (c *Redis) InsertTxs(ctx context.Context, txs []*types.Transaction) error {
	// Get block index
	var blockIndex int
	if err := c.client.Get(ctx, fmt.Sprintf(KeyBlockByNumber, txs[0].BlockNumber)).Scan(&blockIndex); err != nil {
		return err
	}
	// todo: benchmark with different size of txs
	// todo: this way look quite stupid
	for _, tx := range txs {
		txStr, err := json.Marshal(tx)
		if err != nil {
			c.logger.Debug("cannot marshal txs arr to string")
			return err
		}

		txIndex, err := c.client.LPush(ctx, KeyTxsOfBlockIndex, txStr).Result()
		if err != nil {
			c.logger.Debug("cannot insert txs")
			return err
		}
		txOfBlockIndexByNonce := fmt.Sprintf(KeyTxOfBlockIndexByNonce, blockIndex, tx.Nonce)
		txOfBlockIndexByHash := fmt.Sprintf(KeyTxOfBlockIndexByHash, blockIndex, tx.Hash)
		if err := c.client.Set(ctx, txOfBlockIndexByNonce, txIndex, 0).Err(); err != nil {
			return err
		}
		if err := c.client.Set(ctx, txOfBlockIndexByHash, txIndex, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Redis) TxByHash(ctx context.Context, txHash string) (*types.Transaction, error) {
	panic("implement me")
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber return blockIndex
func (c *Redis) InsertBlock(ctx context.Context, block *types.Block) error {
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
