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

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/explorer-backend/utils"
)

const (
	KeyLatestBlockHeight = "#block#latestHeight"
	KeyLatestBlock       = "#block#latest"

	KeyBlocks            = "#blocks" // List
	KeyBlockHashByHeight = "#block#height#%s#hash"
	KeyTxsOfBlockHeight  = "#block#height#%d#txs"

	KeyLatestStats = "#stats#latest"
	KeyLatestTxs   = "#txs#latest" // List

	KeyErrorBlocks           = "#errorBlocks"           // List
	KeyPersistentErrorBlocks = "#persistentErrorBlocks" // List

	KeyTokenInfo = "#token#info"

	KeyTotalTxs       = "#txs#total"
	KeyTotalHolders   = "#holders#total"
	KeyTotalContracts = "#contracts#total"

	KeyValidatorsList = "#validators" // List
	KeyNodesInfoList  = "#nodesInfo"  // List
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

	return tokenInfo == nil
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

func (c *Redis) ListSize(ctx context.Context, key string) (int64, error) {
	size, err := c.client.LLen(ctx, key).Result()
	if err != nil {
		return -1, err
	}
	if size == 0 {
		return 0, errors.New("cache has no block/transaction")
	}

	return size, nil
}

func (c *Redis) InsertTxsOfBlock(ctx context.Context, block *types.Block) error {
	if len(block.Txs) == 0 {
		return nil
	}
	latestTxsLen, err := c.ListSize(ctx, KeyLatestTxs)
	if latestTxsLen < 0 {
		return err
	}
	blockHeight := block.Height
	keyTxsOfThisBlock := fmt.Sprintf(KeyTxsOfBlockHeight, blockHeight)
	c.logger.Debug("Pushing txs to cache:", zap.String("KeyTxsOfThisBlock", keyTxsOfThisBlock))
	for _, tx := range block.Txs {
		txStr, err := json.Marshal(tx)
		if err != nil {
			c.logger.Debug("cannot marshal txs arr to string")
			return err
		}
		_, err = c.client.LPush(ctx, keyTxsOfThisBlock, txStr).Result()
		if err != nil {
			c.logger.Debug("cannot insert txs")
			return err
		}
		// check if we need to pop old tx from latest transaction list due to max size exceeded
		if latestTxsLen+1 > cfg.LatestTxsLength {
			if err := c.client.LPop(ctx, KeyLatestTxs).Err(); err != nil {
				return err
			}
		}
		// also push to latest transaction list
		if err := c.client.RPush(ctx, KeyLatestTxs, txStr).Err(); err != nil {
			return err
		}
		latestTxsLen++
	}

	// set expiration time for list transaction of this block
	result, err := c.client.Expire(ctx, keyTxsOfThisBlock, cfg.BlockInfoExpTime).Result()
	if err != nil || !result {
		c.logger.Warn("cannot set txs of block expiration time in cache", zap.Bool("result", result), zap.Error(err))
		return err
	}

	c.logger.Debug("Done insert txs to cached")
	return nil
}

func (c *Redis) TxByHash(ctx context.Context, txHash string) (*types.Transaction, error) {
	panic("implement me")
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber
func (c *Redis) InsertBlock(ctx context.Context, block *types.Block) error {
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		c.logger.Debug("cannot get size of #blocks", zap.Error(err))
		return err
	}
	c.logger.Debug("redis block buffer size: ", zap.Int64("size", size), zap.Int64("c.cfg.BlockBuffer", c.cfg.BlockBuffer))
	// Size over buffer then
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
			c.logger.Debug("cannot delete txs of block", zap.Error(err))
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
		c.logger.Debug("Error pushing new block", zap.Error(err))
		return err
	}
	keyBlockHashByHeight := fmt.Sprintf(KeyBlockHashByHeight, block.Hash)
	if err := c.client.Set(ctx, keyBlockHashByHeight, block.Height, cfg.BlockInfoExpTime).Err(); err != nil {
		c.logger.Debug("Error set block height by hash", zap.Error(err))
		return err
	}
	c.logger.Debug("Push new block success", zap.Uint64("height", block.Height))
	if err := c.client.Set(ctx, KeyLatestBlockHeight, block.Height, 0).Err(); err != nil {
		c.logger.Debug("Error set latest block height", zap.Error(err))
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
	len, err := c.client.LLen(ctx, keyTxsOfThisBlock).Result()
	c.logger.Debug("TxsByBlockHeight ", zap.String("keyTxsOfThisBlock", keyTxsOfThisBlock), zap.Int64("len", len), zap.Error(err))
	if err != nil || len == 0 {
		return nil, 0, errors.New("block is not exist in cache")
	}

	var txs []*types.Transaction
	if pagination.Skip > int(len)-1 {
		return txs, uint64(len), nil
	}

	txsStrList, err := c.client.LRange(ctx, keyTxsOfThisBlock, 0, len-1).Result()
	if err != nil {
		return nil, 0, err
	}
	for i := pagination.Skip; i < pagination.Skip+pagination.Limit && i < int(len); i++ {
		var tx *types.Transaction
		err = json.Unmarshal([]byte(txsStrList[i]), &tx)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	return txs, uint64(len), nil
}

func (c *Redis) LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var (
		blockList        []*types.Block
		marshalledBlocks []string
		startIndex       = 0 + int64(pagination.Skip)
		endIndex         = startIndex + int64(pagination.Limit) - 1
	)
	len, err := c.ListSize(ctx, KeyBlocks)
	if err != nil {
		return nil, err
	}
	// return error if startIndex or endIndex is out of cache range, require querying in database instead
	if startIndex >= len || endIndex >= len {
		return nil, errors.New("indexes of latest blocks out of range in cache")
	}

	c.logger.Debug("Getting blocks from cache: ", zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex), zap.Uint64("Current latest block height", c.LatestBlockHeight(ctx)))
	marshalledBlocks, err = c.client.LRange(ctx, KeyBlocks, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	for i := startIndex; i <= endIndex; i++ {
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
	len, err := c.ListSize(ctx, KeyLatestTxs)
	if err != nil {
		return nil, err
	}
	// return error if startIndex or endIndex is out of cache range, require querying in database instead
	if startIndex >= len || endIndex >= len {
		return nil, errors.New("indexes of latest txs out of range in cache")
	}

	c.logger.Debug("Get latest txs from block in cache", zap.String("Key", KeyLatestTxs), zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex))
	marshalledTxs, err = c.client.LRange(ctx, KeyLatestTxs, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	for i := startIndex; i <= endIndex; i++ {
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
	for i := start + 1; i < end; i++ {
		_, err := c.client.LPush(ctx, KeyErrorBlocks, strconv.FormatUint(i, 10)).Result()
		if err != nil {
			return err
		}
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

// Holders summary
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
	c.logger.Debug("TotalHolders", zap.String("Total", result))
	if err != nil {
		// Handle error here
		c.logger.Warn("cannot get total holders values")
	}
	// Convert to int
	totalHolders := utils.StrToUint64(result)

	result, err = c.client.Get(ctx, KeyTotalContracts).Result()
	c.logger.Debug("TotalContracts", zap.String("Total", result))
	if err != nil {
		// Handle error here
		c.logger.Warn("cannot get total holders values")
	}
	// Convert to int
	totalContracts := utils.StrToUint64(result)

	return totalHolders, totalContracts
}

func (c *Redis) Validators(ctx context.Context) ([]*types.Validator, error) {
	valsListLen, err := c.client.LLen(ctx, KeyValidatorsList).Result()
	if err != nil {
		c.logger.Warn("cannot get validators list length from cache")
		return nil, err
	}
	if valsListLen == 0 {
		return nil, nil
	}
	valStrList, err := c.client.LRange(ctx, KeyValidatorsList, 0, valsListLen-1).Result()
	if err != nil {
		c.logger.Warn("cannot get validators list from cache", zap.Int("from", 0), zap.Int64("to", valsListLen-1))
		return nil, err
	}
	var valsList []*types.Validator
	for _, valStr := range valStrList {
		var val *types.Validator
		if err := json.Unmarshal([]byte(valStr), &val); err != nil {
			return nil, err
		}
		valsList = append(valsList, val)
	}
	return valsList, nil
}

func (c *Redis) UpdateValidators(ctx context.Context, vals []*types.Validator) error {
	for _, val := range vals {
		valJSON, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if err := c.client.RPush(ctx, KeyValidatorsList, string(valJSON)).Err(); err != nil {
			c.logger.Warn("cannot push validators to cache")
			return err
		}
	}
	result, err := c.client.Expire(ctx, KeyValidatorsList, cfg.ValidatorsListExpTime).Result()
	if err != nil || !result {
		c.logger.Warn("cannot set validators expiration time in cache", zap.Bool("result", result), zap.Error(err))
		return err
	}
	return nil
}

func (c *Redis) NodesInfo(ctx context.Context) ([]*types.NodeInfo, error) {
	nodesListLen, err := c.client.LLen(ctx, KeyNodesInfoList).Result()
	if err != nil {
		c.logger.Warn("cannot get nodes info length from cache")
		return nil, err
	}
	if nodesListLen == 0 {
		return nil, nil
	}
	nodeStrList, err := c.client.LRange(ctx, KeyNodesInfoList, 0, nodesListLen-1).Result()
	if err != nil {
		c.logger.Warn("cannot get nodes info from cache", zap.Int("from", 0), zap.Int64("to", nodesListLen-1))
		return nil, err
	}
	var nodesList []*types.NodeInfo
	for _, nodeStr := range nodeStrList {
		var node *types.NodeInfo
		if err := json.Unmarshal([]byte(nodeStr), &node); err != nil {
			return nil, err
		}
		nodesList = append(nodesList, node)
	}
	return nodesList, nil
}

func (c *Redis) UpdateNodesInfo(ctx context.Context, nodes []*types.NodeInfo) error {
	for _, node := range nodes {
		nodeJSON, err := json.Marshal(node)
		if err != nil {
			return err
		}
		if err := c.client.RPush(ctx, KeyNodesInfoList, string(nodeJSON)).Err(); err != nil {
			c.logger.Warn("cannot push nodes info to cache")
			return err
		}
	}
	result, err := c.client.Expire(ctx, KeyNodesInfoList, cfg.NodesInfoListExpTime).Result()
	if err != nil || !result {
		c.logger.Warn("cannot set nodes info expiration time in cache", zap.Bool("result", result), zap.Error(err))
		return err
	}
	return nil
}

func (c *Redis) deleteKeysOfBlock(ctx context.Context, block *types.Block) error {
	keys := []string{
		fmt.Sprintf(KeyBlockHashByHeight, block.Hash),
		fmt.Sprintf(KeyTxsOfBlockHeight, block.Height),
	}
	if _, err := c.client.Del(ctx, keys...).Result(); err != nil {
		c.logger.Debug("cannot delete keys", zap.Strings("Keys", keys))
		return err
	}
	c.logger.Debug("deleted block info in cache", zap.Any("height", block.Height))
	return nil
}

func (c *Redis) getBlockInCache(ctx context.Context, height uint64, hash string) (*types.Block, error) {
	len, err := c.ListSize(ctx, KeyBlocks)
	if err != nil {
		return nil, err
	}
	blockStrList, err := c.client.LRange(ctx, KeyBlocks, 0, len-1).Result()
	if err != nil {
		return nil, err
	}
	var block *types.Block
	for _, blockStr := range blockStrList {
		err = json.Unmarshal([]byte(blockStr), &block)
		if err != nil {
			return nil, err
		}
		if (block.Height == height) || (block.Hash == hash) {
			return block, nil
		}
	}
	return nil, errors.New("block not found in cache")
}
