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
	"fmt"
	"strings"
	"time"

	"github.com/kardiachain/go-kardia/lib/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

const (
	cBlocks       = "Blocks"
	cTxs          = "Transactions"
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

//region Blocks

func (m *mongoDB) LatestBlockHeight(ctx context.Context) (uint64, error) {
	latestBlock, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if err != nil || len(latestBlock) == 0 {
		return 0, err
	}
	return latestBlock[0].Height, nil
}

func (m *mongoDB) Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var blocks []*types.Block
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"height": -1}),
		options.Find().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.Find().SetSkip(int64(pagination.Skip)),
		options.Find().SetLimit(int64(pagination.Limit)),
	}

	cursor, err := m.wrapper.C(cBlocks).
		Find(bson.D{}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blocks: %v", err)
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	for cursor.Next(ctx) {
		block := &types.Block{}
		if err := cursor.Decode(&block); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (m *mongoDB) BlockByHeight(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var block types.Block
	if err := m.wrapper.C(cBlocks).FindOne(bson.M{"height": blockNumber},
		options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.FindOne().SetHint(bson.M{"height": -1})).Decode(&block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (m *mongoDB) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	var block types.Block
	err := m.wrapper.C(cBlocks).FindOne(bson.M{"hash": blockHash},
		options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.FindOne().SetHint(bson.M{"hash": 1})).Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (m *mongoDB) IsBlockExist(ctx context.Context, blockHeight uint64) (bool, error) {
	var dbBlock types.Block
	err := m.wrapper.C(cBlocks).FindOne(bson.M{"height": blockHeight}, options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0})).Decode(&dbBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *mongoDB) InsertBlock(ctx context.Context, block *types.Block) error {
	logger := m.logger
	// Upsert block into Blocks
	_, err := m.wrapper.C(cBlocks).Insert(block)
	if err != nil {
		logger.Warn("cannot insert new block", zap.Error(err))
		return fmt.Errorf("cannot insert new block")
	}

	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": block.Height}); err != nil {
		logger.Warn("cannot remove old block txs", zap.Error(err))
		return err
	}

	return nil
}

func (m *mongoDB) DeleteLatestBlock(ctx context.Context) (uint64, error) {
	blocks, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if err != nil {
		m.logger.Warn("cannot get old latest block", zap.Error(err))
		return 0, err
	}
	if len(blocks) == 0 {
		m.logger.Warn("there isn't any block in database now, nothing to delete", zap.Error(err))
		return 0, nil
	}
	if err := m.DeleteBlockByHeight(ctx, blocks[0].Height); err != nil {
		m.logger.Warn("cannot remove old latest block", zap.Error(err), zap.Uint64("latest block height", blocks[0].Height))
		return 0, err
	}
	return blocks[0].Height, nil
}

func (m *mongoDB) DeleteBlockByHeight(ctx context.Context, blockHeight uint64) error {
	if _, err := m.wrapper.C(cBlocks).RemoveAll(bson.M{"height": blockHeight}); err != nil {
		m.logger.Warn("cannot remove old latest block", zap.Error(err), zap.Uint64("latest block height", blockHeight))
		return err
	}
	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": blockHeight}); err != nil {
		m.logger.Warn("cannot remove old latest block txs", zap.Error(err), zap.Uint64("latest block height", blockHeight))
		return err
	}
	return nil
}

func (m *mongoDB) BlocksByProposer(ctx context.Context, proposer string, pagination *types.Pagination) ([]*types.Block, uint64, error) {
	var blocks []*types.Block
	opts := []*options.FindOptions(nil)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.D{{Key: "proposerAddress", Value: 1}, {Key: "time", Value: -1}}),
			options.Find().SetSort(bson.M{"time": -1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}
	cursor, err := m.wrapper.C(cBlocks).
		Find(bson.M{"proposerAddress": proposer}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get txs for block: %v", err)
	}
	for cursor.Next(ctx) {
		block := &types.Block{}
		if err := cursor.Decode(block); err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, block)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cBlocks).Count(bson.M{"proposerAddress": proposer})
	if err != nil {
		return nil, 0, err
	}
	return blocks, uint64(total), nil
}

func (m *mongoDB) CountBlocksOfProposer(ctx context.Context, proposerAddress string) (int64, error) {
	total, err := m.wrapper.C(cBlocks).Count(bson.M{"proposerAddress": proposerAddress})
	if err != nil {
		return 0, err
	}
	return total, nil
}

//endregion Blocks

//region Txs

func (m *mongoDB) TxsCount(ctx context.Context) (uint64, error) {
	total, err := m.wrapper.C(cTxs).Count(bson.M{})
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

func (m *mongoDB) TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error) {
	var txs []*types.Transaction
	opts := []*options.FindOptions(nil)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.M{"blockHash": 1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}
	cursor, err := m.wrapper.C(cTxs).
		Find(bson.M{"blockHash": blockHash}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get txs for block: %v", err)
	}
	for cursor.Next(ctx) {
		tx := &types.Transaction{}
		if err := cursor.Decode(tx); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cTxs).Count(bson.M{"blockHash": blockHash})
	if err != nil {
		return nil, 0, err
	}
	return txs, uint64(total), nil
}

func (m *mongoDB) TxsByBlockHeight(ctx context.Context, blockHeight uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error) {
	var txs []*types.Transaction
	opts := []*options.FindOptions(nil)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.M{"blockNumber": -1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}

	cursor, err := m.wrapper.C(cTxs).
		Find(bson.M{"blockNumber": blockHeight}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get txs for block: %v", err)
	}
	for cursor.Next(ctx) {
		tx := &types.Transaction{}
		if err := cursor.Decode(tx); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cTxs).Count(bson.M{"blockNumber": blockHeight})
	if err != nil {
		return nil, 0, err
	}
	return txs, uint64(total), nil
}

// TxsByAddress return txs match input address in FROM/TO field
func (m *mongoDB) TxsByAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, uint64, error) {
	var txs []*types.Transaction
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.D{{Key: "from", Value: 1}, {Key: "time", Value: -1}}),
		options.Find().SetHint(bson.D{{Key: "to", Value: 1}, {Key: "time", Value: -1}}),
		options.Find().SetSort(bson.M{"time": -1}),
	}
	if pagination != nil {
		opts = append(opts, options.Find().SetSkip(int64(pagination.Skip)), options.Find().SetLimit(int64(pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cTxs).
		Find(bson.M{"$or": []bson.M{{"from": address}, {"to": address}}}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	for cursor.Next(ctx) {
		tx := &types.Transaction{}
		if err := cursor.Decode(tx); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	total, err := m.wrapper.C(cTxs).Count(bson.M{"$or": []bson.M{{"from": address}, {"to": address}}}, nil)
	if err != nil {
		return nil, 0, err
	}

	return txs, uint64(total), nil
}

func (m *mongoDB) TxByHash(ctx context.Context, txHash string) (*types.Transaction, error) {
	var tx *types.Transaction
	err := m.wrapper.C(cTxs).FindOne(bson.M{"hash": txHash}, options.FindOne().SetHint(bson.M{"hash": -1})).Decode(&tx)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tx: %v", err)
	}
	return tx, nil
}

// InsertTxs create bulk writer
func (m *mongoDB) InsertTxs(ctx context.Context, txs []*types.Transaction) error {
	var (
		txsBulkWriter []mongo.WriteModel
	)
	for _, tx := range txs {
		txModel := mongo.NewInsertOneModel().SetDocument(tx)
		txsBulkWriter = append(txsBulkWriter, txModel)
	}
	if len(txsBulkWriter) > 0 {
		if _, err := m.wrapper.C(cTxs).BulkWrite(txsBulkWriter); err != nil {
			return err
		}
	}

	return nil
}

func (m *mongoDB) UpsertTxs(ctx context.Context, txs []*types.Transaction) error {
	var txsBulkWriter []mongo.WriteModel
	for _, tx := range txs {
		txModel := mongo.NewInsertOneModel().SetDocument(tx)
		txsBulkWriter = append(txsBulkWriter, txModel)
	}

	if _, err := m.wrapper.C(cTxs).BulkWrite(txsBulkWriter); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) InsertListTxByAddress(ctx context.Context, list []*types.TransactionByAddress) error {
	var txsBulkWriter []mongo.WriteModel
	for _, txByAddress := range list {
		txByAddressModel := mongo.NewInsertOneModel().SetDocument(txByAddress)
		txsBulkWriter = append(txsBulkWriter, txByAddressModel)
	}

	if _, err := m.wrapper.C(cTxsByAddress).BulkWrite(txsBulkWriter); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) LatestTxs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error) {
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"time": -1}),
		options.Find().SetSort(bson.M{"time": -1}),
		options.Find().SetSkip(int64(pagination.Skip)),
		options.Find().SetLimit(int64(pagination.Limit)),
	}

	var txs []*types.Transaction
	cursor, err := m.wrapper.C(cTxs).Find(bson.D{}, opts...)
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
		tx := &types.Transaction{}
		if err := cursor.Decode(&tx); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

//endregion Txs

//region Address

func (m *mongoDB) AddressByHash(ctx context.Context, address string) (*types.Address, error) {
	var c types.Address
	err := m.wrapper.C(cAddresses).FindOne(bson.M{"address": address}).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %v", err)
	}
	return &c, nil
}

func (m *mongoDB) InsertAddress(ctx context.Context, address *types.Address) error {
	if address.Address != "0x" {
		address.Address = common.HexToAddress(address.Address).String()
	}
	address.BalanceFloat = utils.BalanceToFloat(address.BalanceString)
	_, err := m.wrapper.C(cAddresses).Insert(address)
	if err != nil {
		return err
	}
	return nil
}

// UpdateAddresses update last time those addresses active
func (m *mongoDB) UpdateAddresses(ctx context.Context, addresses []*types.Address) error {
	if addresses == nil || len(addresses) == 0 {
		return nil
	}
	var updateAddressOperations []mongo.WriteModel
	for _, info := range addresses {
		info.Address = common.HexToAddress(info.Address).String()
		info.BalanceFloat = utils.BalanceToFloat(info.BalanceString)
		updateAddressOperations = append(updateAddressOperations,
			mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"address": info.Address}).SetUpdate(bson.M{"$set": info}))
	}
	if _, err := m.wrapper.C(cAddresses).BulkWrite(updateAddressOperations); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) GetTotalAddresses(ctx context.Context) (uint64, uint64, error) {
	totalAddr, err := m.wrapper.C(cAddresses).Count(bson.M{"isContract": false})
	if err != nil {
		return 0, 0, err
	}
	totalContractAddr, err := m.wrapper.C(cAddresses).Count(bson.M{"isContract": true})
	if err != nil {
		return 0, 0, err
	}
	return uint64(totalAddr), uint64(totalContractAddr), nil
}

func (m *mongoDB) GetListAddresses(ctx context.Context, sortDirection int, pagination *types.Pagination) ([]*types.Address, error) {
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"balanceFloat": -1}),
		options.Find().SetSort(bson.M{"balanceFloat": sortDirection}),
		options.Find().SetSkip(int64(pagination.Skip)),
		options.Find().SetLimit(int64(pagination.Limit)),
	}

	var (
		rank  = uint64(pagination.Skip + 1)
		addrs []*types.Address
	)
	cursor, err := m.wrapper.C(cAddresses).Find(bson.D{}, opts...)
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
		addr.Rank = rank
		addrs = append(addrs, addr)
		rank++
	}

	return addrs, nil
}

func (m *mongoDB) Addresses(ctx context.Context) ([]*types.Address, error) {
	var addresses []*types.Address
	cursor, err := m.wrapper.C(cAddresses).Find(bson.M{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cursor.Close(ctx); err != nil {
			return
		}
	}()

	for cursor.Next(ctx) {
		var a types.Address
		if err := cursor.Decode(&a); err != nil {
			return nil, err
		}
		addresses = append(addresses, &a)
	}
	return addresses, nil
}

func (m *mongoDB) GetAddressInfo(ctx context.Context, hash string) (*types.Address, error) {
	var address *types.Address
	if err := m.wrapper.C(cAddresses).FindOne(bson.M{"address": hash}).Decode(&address); err != nil {
		return nil, err
	}

	return address, nil
}

//endregion Address

// start region Proposal

func (m *mongoDB) AddVoteToProposal(ctx context.Context, proposalInfo *types.ProposalDetail, voteOption uint64) error {
	m.logger.Warn("AddVoteToProposal", zap.Any("proposal", proposalInfo))
	currentProposal, _ := m.ProposalInfo(ctx, proposalInfo.ID)
	if currentProposal == nil {
		currentProposal = proposalInfo
	}
	// update number of vote choices
	switch voteOption {
	case 0:
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain + 1
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	case 1:
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes + 1
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	case 2:
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo + 1
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
	}
	if err := m.upsertProposal(proposalInfo); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertProposal(ctx context.Context, proposalInfo *types.ProposalDetail) error {
	m.logger.Warn("UpsertProposal", zap.Any("proposal", proposalInfo))
	currentProposal, _ := m.ProposalInfo(ctx, proposalInfo.ID)
	if currentProposal != nil { // need to update these stats from db first
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	}
	if err := m.upsertProposal(proposalInfo); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) ProposalInfo(ctx context.Context, proposalID uint64) (*types.ProposalDetail, error) {
	var result *types.ProposalDetail
	err := m.wrapper.C(cProposal).FindOne(bson.M{"id": proposalID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mongoDB) upsertProposal(proposalInfo *types.ProposalDetail) error {
	proposalInfo.UpdateTime = time.Now().Unix()
	m.logger.Warn("upsertProposal", zap.Any("proposal", proposalInfo))
	model := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"id": proposalInfo.ID}).SetUpdate(bson.M{"$set": proposalInfo}).SetHint(bson.M{"id": -1}),
	}
	if _, err := m.wrapper.C(cProposal).BulkWrite(model); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) GetListProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalDetail, uint64, error) {
	var (
		opts      []*options.FindOptions
		proposals []*types.ProposalDetail
	)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.M{"id": -1}),
			options.Find().SetSort(bson.M{"id": 1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}
	cursor, err := m.wrapper.C(cProposal).
		Find(bson.M{}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get list proposals: %v", err)
	}
	for cursor.Next(ctx) {
		proposal := &types.ProposalDetail{}
		if err := cursor.Decode(proposal); err != nil {
			return nil, 0, err
		}
		proposals = append(proposals, proposal)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cProposal).Count(bson.M{})
	if err != nil {
		return nil, 0, err
	}
	return proposals, uint64(total), nil
}

// end region Proposal

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
