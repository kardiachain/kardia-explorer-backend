// Package server
package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/explorer-backend/utils"
)

type InfoServer interface {
	// API
	LatestBlockHeight(ctx context.Context) (uint64, error)

	// DB
	LatestInsertBlockHeight(ctx context.Context) (uint64, error)

	// Share interface
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)

	ImportBlock(ctx context.Context, block *types.Block) (*types.Block, error)
}

// infoServer handle how data was retrieved, stored without interact with other network excluded dbClient
type infoServer struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	metrics *metrics.Provider

	logger *zap.Logger
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

// ImportBlock handle workflow of import block into system
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block) error {
	// Update cacheClient with simple struct for tracking
	if err := s.cacheClient.InsertBlock(ctx, block); err != nil {
		s.logger.Debug("cannot import block to cache", zap.Error(err))
	}

	// Start import block
	// consider new routine here
	// todo: add metrics
	// todo @longnd: Use redis or leveldb as mem-write buffer for N blocks
	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		s.logger.Debug("cannot import block to db", zap.Error(err))
		return err
	}

	if err := s.cacheClient.InsertTxs(ctx, block.Txs); err != nil {
		s.logger.Debug("cannot import txs to cache", zap.Error(err))
		return err
	}

	insertTxTime := time.Now()
	// todo @trinh: Consider using avgTime metric for reporting/monitor
	if err := s.dbClient.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}
	insertTxConsume := time.Since(insertTxTime)
	s.logger.Debug("Total time for import tx", zap.Any("TimeConsumed", insertTxConsume))

	insertReceiptsTime := time.Now()
	if err := s.ImportReceipts(ctx, block); err != nil {
		return err
	}
	insertReceiptsConsume := time.Since(insertReceiptsTime)
	s.logger.Debug("Total time for import receipt", zap.Any("TimeConsumed", insertReceiptsConsume))

	return nil
}

func (s *infoServer) ImportReceipts(ctx context.Context, block *types.Block) error {
	var listTxByFromAddress []*types.TransactionByAddress
	var listTxByToAddress []*types.TransactionByAddress
	jobs := make(chan types.Transaction, block.NumTxs)
	type response struct {
		err         error
		txByFromAdd *types.TransactionByAddress
		txByToAdd   *types.TransactionByAddress
	}
	results := make(chan response, block.NumTxs)

	//todo @longnd: Move this workers to config or dynamic settings
	for w := 0; w <= 10; w++ {
		go func(jobs <-chan types.Transaction, results chan<- response) {
			for tx := range jobs {
				//s.logger.Debug("Start worker", zap.Any("TX", tx))
				receipt, err := s.kaiClient.GetTransactionReceipt(ctx, tx.Hash)
				if err != nil {
					s.logger.Warn("get receipt err", zap.String("tx hash", tx.Hash), zap.Error(err))
					//todo: consider how we handle this err, just skip it now
					results <- response{
						err: err,
					}
					continue
				}
				toAddress := tx.To
				if tx.To == "" {
					if !utils.IsNilAddress(receipt.ContractAddress) {
						tx.ContractAddress = receipt.ContractAddress
					}
					tx.Status = receipt.Status == 1
					toAddress = tx.ContractAddress
				}

				address, err := s.dbClient.AddressByHash(ctx, toAddress)
				if err != nil {
					//todo: consider how we handle this err, just skip it now
					s.logger.Warn("cannot get address by hash")
					results <- response{
						err: err,
					}
					continue
				}

				if address == nil || address.IsContract {
					var addresses []string
					for _, l := range receipt.Logs {
						addresses = append(addresses, l.Address)
					}

					if err := s.dbClient.UpdateActiveAddresses(ctx, addresses); err != nil {
						//todo: consider how we handle this err, just skip it now
						s.logger.Warn("cannot update active address")
						results <- response{
							err: err,
						}
						continue
					}
				}
				var res response
				res.txByFromAdd = &types.TransactionByAddress{
					Address: tx.From,
					TxHash:  tx.Hash,
					Time:    tx.Time,
				}

				if tx.From != toAddress {
					res.txByToAdd = &types.TransactionByAddress{
						Address: toAddress,
						TxHash:  tx.Hash,
						Time:    tx.Time,
					}
				}
				results <- res
			}
		}(jobs, results)
	}

	for _, tx := range block.Txs {
		jobs <- *tx
	}
	close(jobs)
	// todo @longnd: try to remove this loop
	size := int(block.NumTxs)
	for i := 0; i < size; i++ {
		r := <-results
		if r.err != nil {
			continue
		}
		if r.txByFromAdd != nil {
			listTxByFromAddress = append(listTxByFromAddress, r.txByFromAdd)
		}
		if r.txByToAdd != nil {
			listTxByToAddress = append(listTxByToAddress, r.txByToAdd)
		}
	}

	// todo @longnd: Handle insert failed
	if len(listTxByToAddress) > 0 {
		s.logger.Debug("ListTxFromAddress", zap.Int("Size", len(listTxByFromAddress)))
		if err := s.dbClient.InsertListTxByAddress(ctx, listTxByFromAddress); err != nil {
			return err
		}
	}

	// todo @longnd: Handle insert failed
	if len(listTxByToAddress) > 0 {
		s.logger.Debug("ListTxByToAddress", zap.Int("Size", len(listTxByFromAddress)))
		if err := s.dbClient.InsertListTxByAddress(ctx, listTxByToAddress); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBlock make a simple cache for block
type ValidateBlockStrategy func(db, network *types.Block) bool

// ValidateBlock called by backfill
func (s *infoServer) ValidateBlock(ctx context.Context, block *types.Block, validator ValidateBlockStrategy) error {
	networkBlock, err := s.kaiClient.BlockByHeight(ctx, block.Height)
	if err != nil {
		s.logger.Warn("cannot fetch block from network", zap.Uint64("height", block.Height))
		return err
	}

	isBlockImported, err := s.dbClient.IsBlockExist(ctx, block)
	if err != nil || !isBlockImported {
		if err := s.dbClient.InsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("cannot import block", zap.String("bHash", block.Hash))
			return err
		}
	}

	dbBlock, err := s.dbClient.BlockByHeight(ctx, block.Height)
	if err != nil || !validator(dbBlock, networkBlock) {
		// Force dbBlock with new information from network block
		if err := s.dbClient.UpsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("cannot import block", zap.String("bHash", block.Hash))
			return err
		}
	}

	return nil
}

// calculateTPS return TPS per each [10, 20, 50] blocks
func (s *infoServer) calculateTPS(startTime uint64) (uint64, error) {
	return 0, nil
}

// getAddressByHash return *types.Address from mgo.Collection("Address")
func (s *infoServer) getAddressByHash(address string) (*types.Address, error) {
	return nil, nil
}

func (s *infoServer) getTxsByBlockNumber(blockNumber int64, filter *types.Pagination) ([]*types.Transaction, error) {
	return nil, nil
}
