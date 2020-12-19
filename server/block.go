// Package server
package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Block interface {
	Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)
	LatestBlockHeight(ctx context.Context) (uint64, error)
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	BlockTxs(ctx context.Context, block *types.Block, pagination *types.Pagination) ([]*types.Transaction, error)

	BlockByHeightFromRPC(ctx context.Context, blockHeight uint64) (*types.Block, error)

	ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error
	DeleteLatestBlock(ctx context.Context) (uint64, error)
	DeleteBlockByHeight(ctx context.Context, height uint64) error
	UpsertBlock(ctx context.Context, block *types.Block) error

	InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error
	PopErrorBlockHeight(ctx context.Context) (uint64, error)
	InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error
	InsertUnverifiedBlocks(ctx context.Context, height uint64) error
	PopUnverifiedBlockHeight(ctx context.Context) (uint64, error)

	VerifyBlock(ctx context.Context, blockHeight uint64, networkBlock *types.Block) (bool, error)
}

func (s *infoServer) Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	cBlocks, err := s.cacheClient.LatestBlocks(ctx, pagination)
	if err == nil {
		return cBlocks, nil
	}

	dbBlocks, err := s.dbClient.Blocks(ctx, pagination)
	if err != nil {
		return nil, err
	}
	return dbBlocks, nil
}

// BlockByHash return block by its hash
func (s *infoServer) BlockByHash(ctx context.Context, hash string) (*types.Block, error) {
	lgr := s.logger.With(zap.String("Hash", hash))
	cacheBlock, err := s.cacheClient.BlockByHash(ctx, hash)
	if err == nil {
		return cacheBlock, nil
	}
	lgr.Debug("cannot find block in cache", zap.Error(err))

	dbBlock, err := s.dbClient.BlockByHash(ctx, hash)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Warn("cannot find block in db", zap.Error(err))
	return s.kaiClient.BlockByHash(ctx, hash)
}

// BlockByHeight return a block by height number
// If our network got stuck atm, then make request to mainnet
func (s *infoServer) BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	lgr := s.logger.With(zap.Uint64("Height", blockHeight))
	cacheBlock, err := s.cacheClient.BlockByHeight(ctx, blockHeight)
	if err == nil {
		return cacheBlock, nil
	}
	lgr.Debug("cannot find block in cache")

	dbBlock, err := s.dbClient.BlockByHeight(ctx, blockHeight)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Warn("cannot find block in db")

	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// BlockByHeightFromRPC get block from RPC based on block number
func (s *infoServer) BlockByHeightFromRPC(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

func (s *infoServer) BlockTxs(ctx context.Context, block *types.Block, pagination *types.Pagination) ([]*types.Transaction, error) {
	cTxs, _, err := s.cacheClient.TxsByBlockHash(ctx, block.Hash, pagination)
	if err == nil {
		return cTxs, nil
	}

	dbTxs, _, err := s.dbClient.TxsByBlockHash(ctx, block.Hash, pagination)
	if err == nil {
		return dbTxs, nil
	}

	blockRPC, err := s.kaiClient.BlockByHash(ctx, block.Hash)
	if err != nil {
		return nil, err
	}

	nTxs := blockRPC.Txs
	if pagination.Skip > len(nTxs) {
		nTxs = []*types.Transaction(nil)
	} else if pagination.Skip+pagination.Limit > len(nTxs) {
		nTxs = blockRPC.Txs[pagination.Skip:len(nTxs)]
	} else {
		nTxs = blockRPC.Txs[pagination.Skip : pagination.Skip+pagination.Limit]
	}

	return nTxs, nil
}

// ImportBlock handle workflow of import block into system
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error {
	s.logger.Info("Importing block:", zap.Uint64("Height", block.Height),
		zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if isExist, err := s.dbClient.IsBlockExist(ctx, block.Height); err != nil || isExist {
		return types.ErrRecordExist
	}

	if writeToCache {
		if err := s.cacheClient.InsertBlock(ctx, block); err != nil {
			s.logger.Debug("cannot import block to cache", zap.Error(err))
		}
	}

	// merge receipts into corresponding transactions
	// because getBlockByHash/Height API returns 2 array contains txs and receipts separately
	block.Txs = mergeReceipts(block.Txs, block.Receipts)

	// Start import block
	// consider new routine here
	// todo: add metrics
	// todo @longnd: Use redis or leveldb as mem-write buffer for N blocks
	startTime := time.Now()
	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		s.logger.Debug("cannot import block to db", zap.Error(err))
		return err
	}
	endTime := time.Since(startTime)
	s.metrics.RecordInsertBlockTime(endTime)
	s.logger.Debug("Total time for import block", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertBlockTime()))

	if writeToCache {
		if err := s.cacheClient.InsertTxsOfBlock(ctx, block); err != nil {
			s.logger.Debug("cannot import txs to cache", zap.Error(err))
			return err
		}
	}

	startTime = time.Now()
	if err := s.dbClient.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}
	endTime = time.Since(startTime)
	s.metrics.RecordInsertTxsTime(endTime)
	s.logger.Debug("Total time for import tx", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertTxsTime()))

	// update active addresses
	startTime = time.Now()
	addrList, contractList := filterAddrSet(block.Txs)
	if err := s.dbClient.UpdateActiveAddresses(ctx, addrList, contractList); err != nil {
		return err
	}

	endTime = time.Since(startTime)
	s.metrics.RecordInsertActiveAddressTime(endTime)
	s.logger.Debug("Total time for import active addresses", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertActiveAddressTime()))
	startTime = time.Now()
	totalAddr, totalContractAddr, err := s.dbClient.GetTotalActiveAddresses(ctx)
	if err != nil {
		return err
	}
	err = s.cacheClient.UpdateTotalHolders(ctx, totalAddr, totalContractAddr)
	if err != nil {
		return err
	}
	s.logger.Debug("Total time for getting active addresses", zap.Duration("TimeConsumed", time.Since(startTime)))

	if _, err := s.cacheClient.UpdateTotalTxs(ctx, block.NumTxs); err != nil {
		return err
	}
	return nil
}

func (s *infoServer) DeleteLatestBlock(ctx context.Context) (uint64, error) {
	height, err := s.dbClient.DeleteLatestBlock(ctx)
	if err != nil {
		s.logger.Warn("cannot remove old latest block", zap.Error(err))
		return 0, err
	}
	return height, nil
}

func (s *infoServer) DeleteBlockByHeight(ctx context.Context, height uint64) error {
	err := s.dbClient.DeleteBlockByHeight(ctx, height)
	if err != nil {
		s.logger.Warn("cannot remove block in database by height", zap.Error(err))
		return err
	}
	return nil
}

func (s *infoServer) UpsertBlock(ctx context.Context, block *types.Block) error {
	s.logger.Info("Upserting block:", zap.Uint64("Height", block.Height), zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if err := s.dbClient.DeleteBlockByHeight(ctx, block.Height); err != nil {
		return err
	}
	return s.ImportBlock(ctx, block, false)
}

func (s *infoServer) InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error {
	err := s.cacheClient.InsertErrorBlocks(ctx, start, end)
	if err != nil {
		s.logger.Warn("Cannot insert error block into retry list", zap.Uint64("Start", start), zap.Uint64("End", end))
		return err
	}
	return nil
}

func (s *infoServer) PopErrorBlockHeight(ctx context.Context) (uint64, error) {
	height, err := s.cacheClient.PopErrorBlockHeight(ctx)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (s *infoServer) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return s.kaiClient.LatestBlockNumber(ctx)
}
