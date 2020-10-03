// Package server
package server

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/types"
)

const (
	poolLimit int = 128
)

// DB define list API used by infoServer
type DB interface {
	ping() error
	importBlock(ctx context.Context, block *types.Block) error
	updateActiveAddress() error
}

type InfoServer interface {
	BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error)
	ImportBlock(ctx context.Context, block *types.Block) (*types.Block, error)
}

// infoServer handle how data was retrieved, stored without interact with other network excluded db
type infoServer struct {
	db      DB
	metrics *metrics.Provider
	logger  *zap.Logger
	cache   *redis.Client
}

func (s *infoServer) BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	panic("implement me")
}

func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block) error {
	return s.db.importBlock(ctx, block)
}

func (s *infoServer) createMongoClient() {

}

func (s *infoServer) pingDB() {

}

func (s *infoServer) importBlock(ctx context.Context, block *types.Block) (*types.Block, error) {
	lgr := s.logger
	lgr.Info("Import block")

	// Upsert block with id = height

	// Remove txs belong to this block if any exist

	// Process block's txs
	for _, tx := range block.Txs {
		//todo @longnd: consider put log here since its may delay if not handled careful
		lgr.Debug("process txs")
		if tx.To == "" {

		}

	}

	return nil, nil
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

// getLatestBlock return 50 latest block from cache
func (s *infoServer) getLatestBlock() ([]*types.Block, error) {
	var blocks []*types.Block
	if err := s.cache.Get(KeyLatestBlocks).Scan(&blocks); err != nil {
		// Query latest blocks
	}
	return blocks, nil
}

func (s *infoServer) getStats() (*types.Stats, error) {
	var stats *types.Stats
	if err := s.cache.Get(KeyLatestStats).Scan(stats); err != nil {
		// Query from `stats` collection
	}
	return stats, nil
}

// insertStats insert new stats record for each 24h
func (s *infoServer) insertStats() (*types.Stats, error) {
	stats := &types.Stats{
		UpdatedAt: time.Now(),
	}

	if err := s.cache.Set(KeyLatestStats, stats, DefaultExpiredTime).Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
