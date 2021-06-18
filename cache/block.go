// Package cache
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

type IBlock interface {
	InsertBlock(ctx context.Context, block *types.Block) error
	InsertTxsOfBlock(ctx context.Context, block *types.Block) error
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHeight(ctx context.Context, blockHeight uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error)

	ListSize(ctx context.Context, key string) (int64, error)

	LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)
	LatestTransactions(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)

	InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error
	PopErrorBlockHeight(ctx context.Context) (uint64, error)
	InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error
	PersistentErrorBlockHeights(ctx context.Context) ([]uint64, error)
	InsertUnverifiedBlocks(ctx context.Context, height uint64) error
	PopUnverifiedBlockHeight(ctx context.Context) (uint64, error)
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber
func (c *Redis) InsertBlock(ctx context.Context, block *types.Block) error {
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		return err
	}
	// Size over buffer then
	if size >= c.cfg.BlockBuffer && size != 0 {
		// Get last
		var blockStr string
		if err := c.client.RPop(ctx, KeyBlocks).Scan(&blockStr); err != nil {
			return err
		}

		var block types.Block
		if err := json.Unmarshal([]byte(blockStr), &block); err != nil {
			return err
		}

		if err := c.deleteKeysOfBlock(ctx, &block); err != nil {
			return err
		}
	}

	// Push new block to list
	// Using clone instead assign
	blockCache := *block
	blockCache.Txs = []*types.Transaction{}
	blockCache.Receipts = []*types.Receipt{}

	// marshal current block
	blockJSON, err := json.Marshal(blockCache)
	if err != nil {
		return err
	}
	// Push to top
	_, err = c.client.LPush(ctx, KeyBlocks, blockJSON).Result()
	if err != nil {
		return err
	}
	keyBlockHashByHeight := fmt.Sprintf(KeyBlockHashByHeight, block.Hash)
	if err := c.client.Set(ctx, keyBlockHashByHeight, block.Height, cfg.BlockInfoExpTime).Err(); err != nil {
		return err
	}
	if err := c.client.Set(ctx, KeyLatestBlockHeight, block.Height, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (c *Redis) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	return c.getBlockInCache(ctx, 0, blockHash)
}

func (c *Redis) BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	return c.getBlockInCache(ctx, blockHeight, "")
}

func (c *Redis) TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error) {
	keyBlockHashByHeight := fmt.Sprintf(KeyBlockHashByHeight, blockHash)
	heightStr, err := c.client.Get(ctx, keyBlockHashByHeight).Result()
	if err != nil {
		return nil, 0, errors.New("block not found in cache")
	}
	height := utils.StrToUint64(heightStr)
	return c.TxsByBlockHeight(ctx, height, pagination)
}

func (c *Redis) TxsByBlockHeight(ctx context.Context, blockHeight uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error) {
	keyTxsOfThisBlock := fmt.Sprintf(KeyTxsOfBlockHeight, blockHeight)
	length, err := c.client.LLen(ctx, keyTxsOfThisBlock).Result()
	if err != nil || length == 0 {
		return nil, 0, errors.New("block is not exist in cache")
	}

	var txs []*types.Transaction
	if pagination.Skip > int(length)-1 {
		return txs, uint64(length), nil
	}

	txsStrList, err := c.client.LRange(ctx, keyTxsOfThisBlock, 0, length-1).Result()
	if err != nil {
		return nil, 0, err
	}
	for i := pagination.Skip; i < pagination.Skip+pagination.Limit && i < int(length); i++ {
		var tx *types.Transaction
		err = json.Unmarshal([]byte(txsStrList[i]), &tx)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	return txs, uint64(length), nil
}

func (c *Redis) LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var (
		blockList        []*types.Block
		marshalledBlocks []string
		startIndex       = 0 + int64(pagination.Skip)
		endIndex         = startIndex + int64(pagination.Limit) - 1
	)
	length, err := c.ListSize(ctx, KeyBlocks)
	if err != nil {
		return nil, err
	}
	// return error if startIndex or endIndex is out of cache range, require querying in database instead
	if startIndex >= length || endIndex >= length {
		return nil, errors.New("indexes of latest blocks out of range in cache")
	}

	marshalledBlocks, err = c.client.LRange(ctx, KeyBlocks, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(marshalledBlocks); i++ {
		var b types.Block
		err := json.Unmarshal([]byte(marshalledBlocks[i]), &b)
		if err != nil {
			return nil, err
		}
		blockList = append(blockList, &b)
	}
	return blockList, nil
}

func (c *Redis) LatestTransactions(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error) {
	var (
		txList        []*types.Transaction
		marshalledTxs []string
		startIndex    = 0 + int64(pagination.Skip)
		endIndex      = startIndex + int64(pagination.Limit) - 1
	)
	length, err := c.ListSize(ctx, KeyLatestTxs)
	if err != nil {
		return nil, err
	}
	// return error if startIndex or endIndex is out of cache range, require querying in database instead
	if startIndex >= length || endIndex >= length {
		return nil, errors.New("indexes of latest txs out of range in cache")
	}
	marshalledTxs, err = c.client.LRange(ctx, KeyLatestTxs, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(marshalledTxs); i++ {
		var tx types.Transaction
		err := json.Unmarshal([]byte(marshalledTxs[i]), &tx)
		if err != nil {
			return nil, err
		}
		txList = append(txList, &tx)
	}
	return txList, nil
}

func (c *Redis) InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error {
	var missedBlocks []interface{}
	for i := start + 1; i < end; i++ {
		missedBlocks = append(missedBlocks, strconv.FormatUint(i, 10))
	}
	if len(missedBlocks) == 0 {
		return nil
	}
	_, err := c.client.LPush(ctx, KeyErrorBlocks, missedBlocks...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PopErrorBlockHeight(ctx context.Context) (uint64, error) {
	heightStr, err := c.client.LPop(ctx, KeyErrorBlocks).Result()
	if err != nil {
		return 0, err
	}
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Redis) InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error {
	_, err := c.client.RPush(ctx, KeyPersistentErrorBlocks, strconv.FormatUint(blockHeight, 10)).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PersistentErrorBlockHeights(ctx context.Context) ([]uint64, error) {
	heightsStr, err := c.client.LRange(ctx, KeyPersistentErrorBlocks, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var heights []uint64
	for _, str := range heightsStr {
		height, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, err
		}
		heights = append(heights, height)
	}
	return heights, nil
}

func (c *Redis) InsertUnverifiedBlocks(ctx context.Context, height uint64) error {
	err := c.client.LPush(ctx, KeyUnverifiedBlocks, strconv.FormatUint(height, 10)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PopUnverifiedBlockHeight(ctx context.Context) (uint64, error) {
	heightStr, err := c.client.RPop(ctx, KeyUnverifiedBlocks).Result()
	if err != nil {
		return 0, err
	}
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}
