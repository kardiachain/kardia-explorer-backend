// Package server
package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
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
	lgr.Debug("cannot find in db")

	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// ImportBlock make a simple cache for block
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block) error {
	// Update cacheClient with simple struct for tracking
	if err := s.cacheClient.ImportBlock(ctx, block); err != nil {
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

	// todo: handle receipts


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

// getLatestBlock return 50 latest block from cacheClient
func (s *infoServer) getLatestBlock(ctx context.Context) ([]*types.Block, error) {
	var blocks []*types.Block

	return blocks, nil
}
