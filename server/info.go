// Package server
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
)

type InfoServer interface {
	BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error)
	ImportBlock(ctx context.Context, block *types.Block) (*types.Block, error)
}

// infoServer handle how data was retrieved, stored without interact with other network excluded dbClient
type infoServer struct {
	dbClient    db.Client
	cacheClient *redis.Client
	kaiClient   kardia.ClientInterface

	metrics *metrics.Provider

	logger *zap.Logger
}

// BlockByNumber return a block by height number
// If our network got stuck atm, then make request to mainnet
func (s *infoServer) BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var latestBlockNumber uint64
	if err := s.cacheClient.Get(ctx, KeyLatestBlock).Scan(latestBlockNumber); err != nil || latestBlockNumber < blockNumber {
		// If error then we assume we got some problems with our system
		// or our explorer is behind with mainnet
		// then make request to our mainnet instead waiting for network call
		s.logger.Warn("Delay with latest block")
		return s.kaiClient.BlockByNumber(ctx, blockNumber)
	}

	keyBlockByNumber := fmt.Sprintf(KeyBlockByNumber, blockNumber)
	cacheBlock := &types.CacheBlock{}
	if err := s.cacheClient.Get(ctx, keyBlockByNumber).Scan(&cacheBlock); err != nil {
		return s.kaiClient.BlockByNumber(ctx, blockNumber)
	}

	if !cacheBlock.IsSynced {
		return s.kaiClient.BlockByNumber(ctx, blockNumber)
	}

	return s.dbClient.BlockByNumber(ctx, blockNumber)
}

// ImportBlock make a simple cache for block
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block) error {
	// Update cacheClient with simple struct for tracking
	keyBlockByNumber := fmt.Sprintf(KeyBlockByNumber, block.Height)
	keyBlockByHash := fmt.Sprintf(KeyBlockByHash, block.BlockHash)
	cacheBlock := &types.CacheBlock{
		Hash:     block.BlockHash,
		Number:   block.Height,
		IsSynced: false,
	}
	if _, err := s.cacheClient.Set(ctx, keyBlockByNumber, cacheBlock, 0).Result(); err != nil {
		//
	}
	if _, err := s.cacheClient.Set(ctx, keyBlockByHash, cacheBlock, 0).Result(); err != nil {
		//
	}

	// Start import block
	// todo @longnd: Use redis or leveldb as mem-write buffer for N blocks
	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		return err
	}

	cacheBlock.IsSynced = true
	if _, err := s.cacheClient.Set(ctx, keyBlockByNumber, cacheBlock, 0).Result(); err != nil {
		//
	}
	if _, err := s.cacheClient.Set(ctx, keyBlockByHash, cacheBlock, 0).Result(); err != nil {
		//
	}

	return nil
}

// ValidateBlock make a simple cache for block
type ValidateBlockStrategy func(db, network *types.Block) bool

func (s *infoServer) ValidateBlock(ctx context.Context, block *types.Block, validator ValidateBlockStrategy) error {
	networkBlock, err := s.kaiClient.BlockByNumber(ctx, block.Height)
	if err != nil {
		s.logger.Warn("cannot fetch block from network", zap.Uint64("height", block.Height))
		return err
	}

	isBlockImported, err := s.dbClient.IsBlockExist(ctx, block)
	if err != nil || !isBlockImported {
		if err := s.dbClient.InsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("cannot import block", zap.String("bHash", block.BlockHash))
			return err
		}
	}

	dbBlock, err := s.dbClient.BlockByNumber(ctx, block.Height)
	if err != nil || !validator(dbBlock, networkBlock) {
		// Force dbBlock with new information from network block
		if err := s.dbClient.UpsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("cannot import block", zap.String("bHash", block.BlockHash))
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

func (s *infoServer) getTxsByBlockNumber(blockNumber int64, filter *types.PaginationFilter) ([]*types.Transaction, error) {
	return nil, nil
}

// getLatestBlock return 50 latest block from cacheClient
func (s *infoServer) getLatestBlock(ctx context.Context) ([]*types.Block, error) {
	var blocks []*types.Block
	if err := s.cacheClient.Get(ctx, KeyLatestBlock).Scan(&blocks); err != nil {
		// Query latest blocks
	}
	return blocks, nil
}

func (s *infoServer) getStats(ctx context.Context) (*types.Stats, error) {
	var stats *types.Stats
	if err := s.cacheClient.Get(ctx, KeyLatestStats).Scan(stats); err != nil {
		// Query from `stats` collection
	}
	return stats, nil
}

// insertStats insert new stats record for each 24h
func (s *infoServer) insertStats(ctx context.Context) (*types.Stats, error) {
	stats := &types.Stats{
		UpdatedAt: time.Now(),
	}

	if err := s.cacheClient.Set(ctx, KeyLatestStats, stats, DefaultExpiredTime).Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
