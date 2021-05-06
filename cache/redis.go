// Package cache
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

const (
	KeyLatestBlockHeight = "#block#latestHeight"

	KeyBlocks                = "#blocks" // List
	KeyBlockHashByHeight     = "#block#height#%s#hash"
	KeyTxsOfBlockHeight      = "#block#height#%d#txs"
	KeyErrorBlocks           = "#blocks#error"           // List
	KeyPersistentErrorBlocks = "#blocks#persistentError" // List
	KeyUnverifiedBlocks      = "#blocks#unverified"      // List

	KeyLatestTxs = "#txs#latest" // List

	KeyTokenInfo     = "#token#info"
	KeySupplyAmounts = "#token#supplies"

	KeyTotalTxs       = "#txs#total"
	KeyTotalHolders   = "#holders#total"
	KeyTotalContracts = "#contracts#total"

	KeyValidatorsList = "#validators"
	KeyNodesInfoList  = "#nodesInfo" // List

	KeyContractABI = "#contracts#abi#%s"

	KeyKRCTokenInfo = "#krc#info#%s"
	KeyAddressInfo  = "#addresses#info#%s"

	KeyServerStatus = "#server#status"
)

type Redis struct {
	cfg    Config
	client *redis.Client

	logger *zap.Logger
}

func (c *Redis) UpdateTokenInfo(ctx context.Context, tokenInfo *types.TokenInfo) error {
	// modify some fields
	supplyInfo, err := c.SupplyAmounts(ctx)
	if err == nil && supplyInfo != nil {
		if supplyInfo.ERC20CirculatingSupply > 0 {
			tokenInfo.ERC20CirculatingSupply = supplyInfo.ERC20CirculatingSupply
		}
		if supplyInfo.MainnetGenesisAmount > 0 {
			tokenInfo.MainnetCirculatingSupply = supplyInfo.MainnetGenesisAmount
		}
	}
	tokenInfo.MarketCap = tokenInfo.Price * float64(tokenInfo.ERC20CirculatingSupply+tokenInfo.MainnetCirculatingSupply)
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
	// get current circulating supply that we updated manually, if exists
	tokenInfo.MarketCap = tokenInfo.Price * float64(tokenInfo.ERC20CirculatingSupply)
	return tokenInfo, nil
}

func (c *Redis) UpdateSupplyAmounts(ctx context.Context, supplyInfo *types.SupplyInfo) error {
	currentSupplyInfo, err := c.SupplyAmounts(ctx)
	if err == nil && currentSupplyInfo != nil {
		if supplyInfo.ERC20CirculatingSupply > 0 {
			currentSupplyInfo.ERC20CirculatingSupply = supplyInfo.ERC20CirculatingSupply
		}
		if supplyInfo.MainnetGenesisAmount > 0 {
			currentSupplyInfo.MainnetGenesisAmount = supplyInfo.MainnetGenesisAmount
		}
	} else {
		currentSupplyInfo = supplyInfo
	}
	data, err := json.Marshal(currentSupplyInfo)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, KeySupplyAmounts, string(data), 0).Err(); err != nil {
		return err
	}
	// then update token info, base on this new supply info
	tokenInfo, err := c.TokenInfo(ctx)
	if err != nil || tokenInfo == nil {
		return nil
	}
	_ = c.UpdateTokenInfo(ctx, tokenInfo)
	return nil
}

func (c *Redis) SupplyAmounts(ctx context.Context) (*types.SupplyInfo, error) {
	result, err := c.client.Get(ctx, KeySupplyAmounts).Result()
	if err != nil {
		return nil, err
	}
	var supplyInfo *types.SupplyInfo
	if err := json.Unmarshal([]byte(result), &supplyInfo); err != nil {
		return nil, err
	}
	return supplyInfo, nil
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
	if err != nil {
		// Handle error here
	}
	// Convert to int
	totalTxs := utils.StrToUint64(result)
	return totalTxs
}

func (c *Redis) UpdateTotalTxs(ctx context.Context, blockTxs uint64) (uint64, error) {
	totalTxs := c.TotalTxs(ctx)
	totalTxs += blockTxs

	if err := c.client.Set(ctx, KeyTotalTxs, totalTxs, 0).Err(); err != nil {
		c.logger.Warn("cannot set total txs values")
	}
	return totalTxs, nil
}

func (c *Redis) SetTotalTxs(ctx context.Context, numTxs uint64) error {
	if err := c.client.Set(ctx, KeyTotalTxs, numTxs, 0).Err(); err != nil {
		c.logger.Warn("cannot set total txs values", zap.Error(err))
		return err
	}
	return nil
}

func (c *Redis) LatestBlockHeight(ctx context.Context) uint64 {
	result, err := c.client.Get(ctx, KeyLatestBlockHeight).Uint64()
	if err != nil {
		// Handle error here
	}
	return result
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
	for _, tx := range block.Txs {
		txStr, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		_, err = c.client.LPush(ctx, keyTxsOfThisBlock, txStr).Result()
		if err != nil {
			return err
		}
		// check if we need to pop old tx from latest transaction list due to max size exceeded
		if latestTxsLen+1 > cfg.LatestTxsLength {
			if err := c.client.RPop(ctx, KeyLatestTxs).Err(); err != nil {
				return err
			}
		}
		// also push to latest transaction list
		if err := c.client.LPush(ctx, KeyLatestTxs, txStr).Err(); err != nil {
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

	return nil
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

// GetListHolders summary
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

func (c *Redis) Validators(ctx context.Context) (*types.Validators, error) {
	validatorsListStr, err := c.client.Get(ctx, KeyValidatorsList).Result()
	if err != nil {
		return nil, err
	}
	if validatorsListStr == "" {
		return nil, errors.New("validators list in cache is empty")
	}
	var validatorsList *types.Validators
	if err := json.Unmarshal([]byte(validatorsListStr), &validatorsList); err != nil {
		return nil, err
	}
	return validatorsList, nil
}

func (c *Redis) UpdateValidators(ctx context.Context, validators *types.Validators) error {
	validatorsJSON, err := json.Marshal(validators)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, KeyValidatorsList, string(validatorsJSON), 15*time.Second).Err(); err != nil {
		c.logger.Warn("cannot set validators list to cache")
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
		return err
	}
	return nil
}

func (c *Redis) getBlockInCache(ctx context.Context, height uint64, hash string) (*types.Block, error) {
	length, err := c.ListSize(ctx, KeyBlocks)
	if err != nil {
		return nil, err
	}
	blockStrList, err := c.client.LRange(ctx, KeyBlocks, 0, length-1).Result()
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

func (c *Redis) SMCAbi(ctx context.Context, key string) (string, error) {
	keyABI := fmt.Sprintf(KeyContractABI, key)
	result, err := c.client.Get(ctx, keyABI).Result()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(result, cfg.SMCTypePrefix) {
		// get the abi of this smc type
		result, err = c.client.Get(ctx, result).Result()
		if err != nil {
			return "", err
		}
	}
	return result, nil
}

func (c *Redis) UpdateSMCAbi(ctx context.Context, key, abi string) error {
	keyABI := fmt.Sprintf(KeyContractABI, key)
	if err := c.client.Set(ctx, keyABI, abi, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) KRCTokenInfo(ctx context.Context, krcTokenAddr string) (*types.KRCTokenInfo, error) {
	keyKRC := fmt.Sprintf(KeyKRCTokenInfo, krcTokenAddr)
	result, err := c.client.Get(ctx, keyKRC).Result()
	if err != nil {
		return nil, err
	}
	var KRCTokenInfo *types.KRCTokenInfo
	if err := json.Unmarshal([]byte(result), &KRCTokenInfo); err != nil {
		return nil, err
	}
	return KRCTokenInfo, nil
}

func (c *Redis) UpdateKRCTokenInfo(ctx context.Context, krcTokenInfo *types.KRCTokenInfo) error {
	keyKRC := fmt.Sprintf(KeyKRCTokenInfo, krcTokenInfo.Address)
	data, err := json.Marshal(krcTokenInfo)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, keyKRC, data, cfg.KRCTokenInfoExpTime).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) AddressInfo(ctx context.Context, addr string) (*types.Address, error) {
	keyAddrInfo := fmt.Sprintf(KeyAddressInfo, addr)
	result, err := c.client.Get(ctx, keyAddrInfo).Result()
	if err != nil {
		return nil, err
	}
	var addrInfo *types.Address
	if err := json.Unmarshal([]byte(result), &addrInfo); err != nil {
		return nil, err
	}
	return addrInfo, nil
}

func (c *Redis) UpdateAddressInfo(ctx context.Context, addrInfo *types.Address) error {
	keyAddrInfo := fmt.Sprintf(KeyAddressInfo, addrInfo.Address)
	data, err := json.Marshal(addrInfo)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, keyAddrInfo, data, cfg.AddressInfoExpTime).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) UpdateServerStatus(ctx context.Context, serverStatus *types.ServerStatus) error {
	data, err := json.Marshal(serverStatus)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, KeyServerStatus, data, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) ServerStatus(ctx context.Context) (*types.ServerStatus, error) {
	result, err := c.client.Get(ctx, KeyServerStatus).Result()
	if err != nil {
		return nil, err
	}
	var serverStatus *types.ServerStatus
	if err := json.Unmarshal([]byte(result), &serverStatus); err != nil {
		return nil, err
	}
	return serverStatus, nil
}
