// Package cache
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/explorer-backend/utils"
)

const (
	KeyLatestBlockHeight = "#block#latestHeight"
	KeyLatestBlock       = "#block#latest"

	KeyBlocks        = "#blocks" // List
	KeyBlockByNumber = "#block#%d"
	KeyBlockByHash   = "#block#%s"

	KeyTxsOfBlockHeight       = "#block#height#%d#txs"
	KeyTxOfBlockHeightByNonce = "#block#height#%d#tx#%d"
	KeyTxOfBlockHeightByHash  = "#block#height#%d#tx#%s"

	KeyLatestStats = "#stats#latest"

	PatternGetAllKeyOfBlockHeight = "#block#height#%d*"

	ErrorBlocks = "#errorBlocks" // List

	KeyTotalTxs = "#txs#total"

	KeyTokenInfo = "#token#info"
)

type Redis struct {
	cfg    Config
	client *redis.Client

	logger *zap.Logger
}

func (c *Redis) UpdateTokenInfo(ctx context.Context, tokenInfo *types.TokenInfo) error {
	data, err := json.Marshal(tokenInfo)
	if err != nil {
		return err
	}

	if _, err := c.client.Set(ctx, KeyTokenInfo, string(data), 60*time.Minute).Result(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) TokenInfo(ctx context.Context) (*types.TokenInfo, error) {
	result, err := c.client.Get(ctx, KeyTokenInfo).Result()
	if err != nil {
		return nil, err
	}
	var tokenInfo *types.TokenInfo
	if err := json.Unmarshal([]byte(result), &tokenInfo); err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

func (c *Redis) IsRequestToCoinMarket(ctx context.Context) bool {
	tokenInfo, err := c.TokenInfo(ctx)
	if err != nil {
		return false
	}

	return tokenInfo != nil
}

func (c *Redis) TotalTxs(ctx context.Context) uint64 {
	result, err := c.client.Get(ctx, KeyTotalTxs).Result()
	c.logger.Debug("TotalTxs", zap.String("Total", result))
	if err != nil {
		// Handle error here
		c.logger.Warn("cannot get total txs values")
	}
	// Convert to int
	totalTxs := utils.StrToUint64(result)
	return totalTxs
}

func (c *Redis) UpdateTotalTxs(ctx context.Context, blockTxs uint64) (uint64, error) {
	totalTxs := c.TotalTxs(ctx)
	totalTxs += blockTxs

	if err := c.client.Set(ctx, KeyTotalTxs, totalTxs, 0).Err(); err != nil {
		// Handle error here
		c.logger.Warn("cannot set total txs values")
	}
	return totalTxs, nil
}

func (c *Redis) LatestBlockHeight(ctx context.Context) uint64 {
	result, err := c.client.Get(ctx, KeyLatestBlockHeight).Uint64()
	c.logger.Debug("LatestBlockHeight", zap.Uint64("Height", result))
	if err != nil {
		// Handle error here
		c.logger.Warn("cannot get latest block height value from cache")
	}
	return result
}

func (c *Redis) PopReceipt(ctx context.Context) (*types.Receipt, error) {
	panic("implement me")
}

func (c *Redis) BlocksSize(ctx context.Context) (int64, error) {
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		return -1, err
	}
	if size == 0 {
		return 0, errors.New("blocks has no cache")
	}

	return size, nil
}

func (c *Redis) InsertTxsOfBlock(ctx context.Context, block *types.Block) error {
	if len(block.Txs) == 0 {
		return nil
	}

	blockHeight := block.Height
	KeyTxsOfThisBlock := fmt.Sprintf(KeyTxsOfBlockHeight, blockHeight)
	c.logger.Debug("Pushing txs to cache:", zap.String("KeyTxsOfThisBlock", KeyTxsOfThisBlock))
	for _, tx := range block.Txs {
		// c.logger.Debug("Tx to cache", zap.Any("Tx", tx))
		txStr, err := json.Marshal(tx)
		if err != nil {
			c.logger.Debug("cannot marshal txs arr to string")
			return err
		}

		txIndex, err := c.client.LPush(ctx, KeyTxsOfThisBlock, txStr).Result()
		if err != nil {
			c.logger.Debug("cannot insert txs")
			return err
		}
		txOfBlockHeightByNonce := fmt.Sprintf(KeyTxOfBlockHeightByNonce, blockHeight, tx.Nonce)
		txOfBlockHeightByHash := fmt.Sprintf(KeyTxOfBlockHeightByHash, blockHeight, tx.Hash)
		if err := c.client.Set(ctx, txOfBlockHeightByNonce, txIndex, c.cfg.DefaultExpiredTime).Err(); err != nil {
			return err
		}
		if err := c.client.Set(ctx, txOfBlockHeightByHash, txIndex, c.cfg.DefaultExpiredTime).Err(); err != nil {
			return err
		}
	}

	c.logger.Debug("Done insert txs to cached")

	return nil
}

func (c *Redis) TxByHash(ctx context.Context, txHash string) (*types.Transaction, error) {
	panic("implement me")
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber return blockIndex
func (c *Redis) InsertBlock(ctx context.Context, block *types.Block) error {
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		c.logger.Debug("cannot get size of #blocks", zap.Error(err))
		return err
	}

	// Size over buffer then
	c.logger.Debug("redis block buffer size: ", zap.Int64("size", size), zap.Int64("c.cfg.BlockBuffer", c.cfg.BlockBuffer))
	if size >= c.cfg.BlockBuffer && size != 0 {
		// Get last
		var blockStr string
		if err := c.client.RPop(ctx, KeyBlocks).Scan(&blockStr); err != nil {
			c.logger.Debug("cannot pop last element", zap.Error(err))
			return err
		}

		var block types.Block
		if err := json.Unmarshal([]byte(blockStr), &block); err != nil {
			c.logger.Debug("cannot marshal block from cache to object", zap.Error(err))
			return err
		}

		if err := c.deleteKeysOfBlock(ctx, &block); err != nil {
			c.logger.Debug("cannot delete keys of block index", zap.Error(err))
			return err
		}
	}

	// Push new block to list
	// Using clone instead assign
	blockCache := *block
	blockCache.Txs = []*types.Transaction{}
	blockCache.Receipts = []*types.Receipt{}

	// Push to top
	lPushResult := c.client.LPush(ctx, KeyBlocks, blockCache.String())
	blockIndex, err := lPushResult.Result() // Always
	if err != nil {
		c.logger.Debug("Error pushing new block", zap.Error(err))
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
	if err := c.client.Set(ctx, KeyLatestBlockHeight, block.Height, 0).Err(); err != nil {
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

func (c *Redis) LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var (
		blockList        []*types.Block
		marshalledBlocks []string
		startIndex       = 0 + int64(pagination.Skip)
		endIndex         = startIndex + int64(pagination.Limit) - 1
	)
	marshalledBlocks, err := c.client.LRange(ctx, KeyBlocks, startIndex, endIndex).Result()
	c.logger.Debug("Getting blocks from cache: ", zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex), zap.Uint64("Current latest block height", c.LatestBlockHeight(ctx)))
	if err != nil {
		return nil, err
	}
	for _, bStr := range marshalledBlocks {
		var b types.Block
		err := json.Unmarshal([]byte(bStr), &b)
		if err != nil {
			return nil, err
		}
		blockList = append(blockList, &b)
	}
	c.logger.Debug("Latest blocks from cache: ", zap.Any("blocks", blockList))
	return blockList, nil
}

func (c *Redis) LatestTransactions(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error) {
	var (
		txList        []*types.Transaction
		marshalledTxs []string
		startIndex    = 0 + int64(pagination.Skip)
		endIndex      = startIndex + int64(pagination.Limit) - 1
	)
	latestBlockHeight := c.LatestBlockHeight(ctx)
	KeyTxsOfLatestBlock := fmt.Sprintf(KeyTxsOfBlockHeight, latestBlockHeight)
	c.logger.Debug("Get latest txs from block", zap.String("Key", KeyTxsOfLatestBlock))
	marshalledTxs, err := c.client.LRange(ctx, KeyTxsOfLatestBlock, startIndex, endIndex).Result()
	c.logger.Debug("Getting txs from cache: ", zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex))
	if err != nil {
		return nil, err
	}

	for _, txStr := range marshalledTxs {
		var tx types.Transaction
		err := json.Unmarshal([]byte(txStr), &tx)
		if err != nil {
			return nil, err
		}
		txList = append(txList, &tx)
	}
	c.logger.Debug("Latest txs from cache: ", zap.Any("txs", txList))
	return txList, nil
}

func (c *Redis) InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error {
	for i := start + 1; i < end; i++ {
		_, err := c.client.LPush(ctx, ErrorBlocks, strconv.FormatUint(i, 10)).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Redis) PopErrorBlockHeight(ctx context.Context) (uint64, error) {
	heightStr, err := c.client.LPop(ctx, ErrorBlocks).Result()
	if err != nil {
		return 0, err
	}
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
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

func (c *Redis) deleteKeysOfBlock(ctx context.Context, block *types.Block) error {
	var keys []string
	pattern := fmt.Sprintf(PatternGetAllKeyOfBlockHeight, block.Height)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Debug("cannot get keys", zap.Error(err))
		return err
	}

	c.logger.Debug("deleting block info in cache", zap.Any("block", block))

	keys = append(keys, []string{
		fmt.Sprintf(KeyBlockByNumber, block.Height),
		fmt.Sprintf(KeyBlockByHash, block.Hash),
	}...)

	if _, err := c.client.Del(ctx, keys...).Result(); err != nil {
		c.logger.Debug("cannot delete keys", zap.Strings("Keys", keys))
		return err
	}

	return nil
}
