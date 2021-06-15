/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
// Package db
package db

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	cBlocks = "Blocks"

	cAddresses    = "Addresses"
	cTxsByAddress = "TransactionsByAddress"
	cStats        = "Stats"
	cProposal     = "Proposal"
)

type mongoDB struct {
	logger  *zap.Logger
	wrapper *KaiMgo
	db      *mongo.Database
}

func newMongoDB(cfg Config) (*mongoDB, error) {

	ctx := context.Background()
	dbClient := &mongoDB{
		logger:  cfg.Logger,
		wrapper: &KaiMgo{},
	}
	mgoOptions := options.Client()
	mgoOptions.ApplyURI(cfg.URL)
	mgoOptions.SetMinPoolSize(uint64(cfg.MinConn))
	mgoOptions.SetMaxPoolSize(uint64(cfg.MaxConn))
	mgoClient, err := mongo.NewClient(mgoOptions)
	if err != nil {
		return nil, err
	}

	if err := mgoClient.Connect(context.Background()); err != nil {
		return nil, err
	}

	dbClient.wrapper.Database(mgoClient.Database(cfg.DbName))

	if cfg.FlushDB {
		cfg.Logger.Info("Start flush database")
		if err := mgoClient.Database(cfg.DbName).Drop(ctx); err != nil {
			return nil, err
		}
	}
	_ = createIndexes(dbClient)

	return dbClient, nil
}

func createIndexes(dbClient *mongoDB) error {
	type CIndex struct {
		c     string
		model []mongo.IndexModel
	}

	indexes := []CIndex{
		// Add index to improve querying transactions by hash, block hash and block height
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.M{"hash": -1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.M{"blockNumber": -1}, Options: options.Index().SetSparse(true)}}},
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.M{"blockHash": 1}, Options: options.Index().SetSparse(true)}}},
		// Add index in `from` and `to` fields to improve get txs of address, considering if memory is increasing rapidly
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.D{{Key: "from", Value: 1}, {Key: "time", Value: -1}}, Options: options.Index().SetSparse(true)}}},
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.D{{Key: "to", Value: 1}, {Key: "time", Value: -1}}, Options: options.Index().SetSparse(true)}}},
		{c: cTxs, model: []mongo.IndexModel{{Keys: bson.M{"time": -1}, Options: options.Index().SetSparse(true)}}},
		// Add index to improve querying blocks by proposer, hash and height
		{c: cBlocks, model: []mongo.IndexModel{{Keys: bson.M{"height": -1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cBlocks, model: []mongo.IndexModel{{Keys: bson.M{"hash": 1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cBlocks, model: []mongo.IndexModel{{Keys: bson.D{{Key: "proposerAddress", Value: 1}, {Key: "time", Value: -1}}, Options: options.Index().SetSparse(true)}}},
		// indexing addresses collection
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"address": 1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"name": 1}, Options: options.Index().SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"isContract": 1}}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"balanceFloat": -1}, Options: options.Index().SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"tokenName": 1}, Options: options.Index().SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"tokenSymbol": 1}, Options: options.Index().SetSparse(true)}}},
		// indexing proposal collection
		{c: cProposal, model: []mongo.IndexModel{{Keys: bson.M{"id": -1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		// indexing validator collection
		{c: cValidators, model: []mongo.IndexModel{{Keys: bson.M{"address": 1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cValidators, model: []mongo.IndexModel{{Keys: bson.M{"name": 1}, Options: options.Index().SetSparse(true)}}},
		// indexing contract & ABI collection
		{c: cContract, model: []mongo.IndexModel{{Keys: bson.M{"name": 1}, Options: options.Index().SetSparse(true)}}},
		{c: cContract, model: []mongo.IndexModel{{Keys: bson.M{"type": 1}, Options: options.Index().SetSparse(true)}}},
		{c: cContract, model: []mongo.IndexModel{{Keys: bson.M{"address": 1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		{c: cABI, model: []mongo.IndexModel{{Keys: bson.M{"type": 1}, Options: options.Index().SetUnique(true).SetSparse(true)}}},
		// indexing contract events collection
		{c: cEvents, model: dbClient.createEventsCollectionIndexes()},
		// indexing token holders collection
		{c: cHolders, model: dbClient.createHoldersCollectionIndexes()},
		// indexing internal txs collection
		{c: cInternalTxs, model: dbClient.createInternalTxsCollectionIndexes()},
		{c: cDelegator, model: createDelegatorCollectionIndexes()},
	}
	for _, cIdx := range indexes {
		if err := dbClient.wrapper.C(cIdx.c).EnsureIndex(cIdx.model); err != nil {
			return err
		}
	}
	return nil
}

//region General

func (m *mongoDB) ping() error {
	return nil
}

func (m *mongoDB) dropCollection(collectionName string) {
	if _, err := m.wrapper.C(collectionName).RemoveAll(nil); err != nil {
		return
	}
}

func (m *mongoDB) dropDatabase(ctx context.Context) error {
	return m.wrapper.DropDatabase(ctx)
}

//endregion General

// region Stats

func (m *mongoDB) UpdateStats(ctx context.Context, stats *types.Stats) error {
	_, err := m.wrapper.C(cStats).Insert(stats)
	if err != nil {
		return err
	}
	// remove old stats
	if _, err := m.wrapper.C(cStats).RemoveAll(bson.M{"updatedAtBlock": bson.M{"$lt": stats.UpdatedAtBlock}}); err != nil {
		m.logger.Warn("cannot remove old stats", zap.Error(err), zap.Uint64("latest updated block", stats.UpdatedAtBlock))
		return err
	}
	return nil
}

func (m *mongoDB) Stats(ctx context.Context) *types.Stats {
	var stats *types.Stats
	if err := m.wrapper.C(cStats).FindOne(bson.M{}).Decode(&stats); err == nil {
		// remove blocks after checkpoint
		latestBlock, err := m.Blocks(ctx, &types.Pagination{
			Skip:  0,
			Limit: 1,
		})
		if len(latestBlock) > 0 {
			stats.UpdatedAtBlock = latestBlock[0].Height
		}
		for {
			if stats.UpdatedAtBlock%cfg.UpdateStatsInterval == 0 {
				break
			}
			stats.UpdatedAtBlock, err = m.DeleteLatestBlock(ctx)
			if err != nil {
				m.logger.Warn("Getting stats: DeleteLatestBlock error", zap.Error(err))
			}
			stats.UpdatedAtBlock--
		}
		return stats
	}
	// create a checkpoint (latestBlockHeight) and remove blocks after checkpoint
	// then calculate stats based on current database
	latestBlockHeight := uint64(0)
	latestBlock, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if len(latestBlock) > 0 {
		latestBlockHeight = latestBlock[0].Height
	}
	for {
		if latestBlockHeight%cfg.UpdateStatsInterval == 0 {
			break
		}
		latestBlockHeight, err = m.DeleteLatestBlock(ctx)
		if err != nil {
			m.logger.Warn("Getting stats: DeleteLatestBlock error", zap.Error(err))
		}
		latestBlockHeight--
	}
	totalAddrs, totalContracts, err := m.GetTotalAddresses(ctx)
	if err != nil {
		totalAddrs = 0
		totalContracts = 0
	}
	totalTxs, err := m.TxsCount(ctx)
	if err != nil {
		totalTxs = 0
	}
	return &types.Stats{
		UpdatedAt:         time.Now(),
		UpdatedAtBlock:    latestBlockHeight,
		TotalTransactions: totalTxs,
		TotalAddresses:    totalAddrs,
		TotalContracts:    totalContracts,
	}
}

// end region Stats

func (m *mongoDB) AddressByName(ctx context.Context, name string) ([]*types.Address, error) {
	var (
		addrs []*types.Address
		opts  = []*options.FindOptions{
			options.Find().SetHint(bson.M{"address": 1}),
			options.Find().SetHint(bson.M{"name": 1}),
		}
	)
	crit := []bson.M{
		{"name": bson.D{{"$regex", primitive.Regex{Pattern: name, Options: "i"}}}},
		{"tokenSymbol": bson.D{{"$regex", primitive.Regex{Pattern: name, Options: "i"}}}},
	}
	if strings.HasPrefix(name, "0x") {
		crit = append(crit, bson.M{"address": bson.D{{"$regex", primitive.Regex{Pattern: name, Options: "i"}}}})
	}
	cursor, err := m.wrapper.C(cAddresses).Find(bson.M{"$or": crit}, opts...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	for cursor.Next(ctx) {
		addr := &types.Address{}
		if err := cursor.Decode(&addr); err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

func (m *mongoDB) ContractByName(ctx context.Context, name string) ([]*types.Contract, error) {
	var (
		contracts []*types.Contract
		opts      = []*options.FindOptions{
			options.Find().SetHint(bson.M{"address": 1}),
			options.Find().SetHint(bson.M{"name": 1}),
		}
	)
	crit := []bson.M{
		{"name": bson.D{{"$regex", primitive.Regex{Pattern: name, Options: "i"}}}},
	}
	if strings.HasPrefix(name, "0x") {
		crit = append(crit, bson.M{"address": bson.D{{"$regex", primitive.Regex{Pattern: name, Options: "i"}}}})
	}
	cursor, err := m.wrapper.C(cContract).Find(bson.M{"$or": crit}, opts...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	for cursor.Next(ctx) {
		smc := &types.Contract{}
		if err := cursor.Decode(&smc); err != nil {
			return nil, err
		}
		contracts = append(contracts, smc)
	}

	return contracts, nil
}
