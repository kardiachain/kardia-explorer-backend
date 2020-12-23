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
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	cBlocks          = "Blocks"
	cTxs             = "Transactions"
	cAddresses       = "Addresses"
	cTxsByAddress    = "TransactionsByAddress"
	cActiveAddresses = "ActiveAddresses"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	hydro             = big.NewInt(1000000000000000000)
)

type mongoDB struct {
	logger  *zap.Logger
	wrapper *KaiMgo
	db      *mongo.Database
}

func newMongoDB(cfg Config) (*mongoDB, error) {
	cfg.Logger.Debug("Create mgo with config", zap.Any("config", cfg))

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
		cfg.Logger.Debug("Start flush database")
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

	for _, cIdx := range []CIndex{
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
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"isContract": 1}}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"balanceFloat": -1}, Options: options.Index().SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"tokenName": 1}, Options: options.Index().SetSparse(true)}}},
		{c: cAddresses, model: []mongo.IndexModel{{Keys: bson.M{"tokenSymbol": 1}, Options: options.Index().SetSparse(true)}}},
	} {
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

//region Blocks

func (m *mongoDB) Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	m.logger.Debug("get blocks from db", zap.Any("pagination", pagination))
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
		//m.logger.Debug("Get block success", zap.Any("block", block))
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
	// todo @longnd: add block info ?
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

// UpsertBlock call by verifier, to avoid duplicate block record
func (m *mongoDB) UpsertBlock(ctx context.Context, block *types.Block) error {
	return ErrNotImplemented
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
	m.logger.Debug("DeleteLatestBlock...", zap.Uint64("latest block height", blocks[0].Height))
	if err := m.DeleteBlockByHeight(ctx, blocks[0].Height); err != nil {
		m.logger.Warn("cannot remove old latest block", zap.Error(err), zap.Uint64("latest block height", blocks[0].Height))
		return 0, err
	}
	m.logger.Debug("DeleteLatestBlock success", zap.Uint64("latest block height", blocks[0].Height))
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

//endregion Blocks

//region Txs

func (m *mongoDB) TxsCount(ctx context.Context) (uint64, error) {
	total, err := m.wrapper.C(cTxs).Count(bson.M{})
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

func (m *mongoDB) BlockTxCount(ctx context.Context, hash string) (int64, error) {
	return 0, nil
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
		options.Find().SetSkip(int64(pagination.Skip)),
		options.Find().SetLimit(int64(pagination.Limit)),
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

func (m *mongoDB) TxByNonce(ctx context.Context, nonce int64) (*types.Transaction, error) {
	var tx *types.Transaction
	err := m.wrapper.C(cTxs).FindOne(bson.M{"nonce": nonce}).Decode(&tx)
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
		// TODO(trinhdn): insert created contract info to database with model `address`
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
		m.logger.Debug("Process tx", zap.String("tx", fmt.Sprintf("%#v", tx)))
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
	start := time.Now()
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
	queryTime := time.Since(start)
	m.logger.Debug("Total time for query tx", zap.Any("TimeConsumed", queryTime))
	for cursor.Next(ctx) {
		tx := &types.Transaction{}
		if err := cursor.Decode(&tx); err != nil {
			return nil, err
		}
		//m.logger.Debug("Get latest txs success", zap.Any("tx", tx))
		txs = append(txs, tx)
	}
	processTime := time.Since(start)
	m.logger.Debug("Total time for process tx", zap.Any("TimeConsumed", processTime))

	return txs, nil
}

//endregion Txs

//region Token
func (m *mongoDB) TokenHolders(ctx context.Context, tokenAddress string, pagination *types.Pagination) ([]*types.TokenHolder, uint64, error) {
	panic("implement me")
}

//endregion Token

//region Address

func (m *mongoDB) AddressByHash(ctx context.Context, addressHash string) (*types.Address, error) {
	var c types.Address
	err := m.wrapper.C(cAddresses).FindOne(bson.M{"address": addressHash}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get address: %v", err)
	}
	return &c, nil
}

func (m *mongoDB) OwnedTokensOfAddress(ctx context.Context, walletAddress string, pagination *types.Pagination) ([]*types.TokenHolder, uint64, error) {
	panic("implement me")
}

// UpdateAddresses update last time those addresses active
// Just skip for now
func (m *mongoDB) UpdateAddresses(ctx context.Context, addressesMap map[string]*big.Int, contractAddrMap map[string]*big.Int) error {
	var updateAddressOperations []mongo.WriteModel
	for addr, balance := range addressesMap {
		updateAddressOperations = append(updateAddressOperations, appendUpdateAddressModels(addr, balance, false))
	}
	for addr, balance := range contractAddrMap {
		updateAddressOperations = append(updateAddressOperations, appendUpdateAddressModels(addr, balance, true))
	}
	if _, err := m.wrapper.C(cAddresses).BulkWrite(updateAddressOperations); err != nil {
		return err
	}
	return nil
}

func appendUpdateAddressModels(addr string, balance *big.Int, isContract bool) mongo.WriteModel {
	balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(hydro)).Float64() //converting to KAI from HYDRO
	balanceString := balance.String()
	updateAddressOperation := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"address": addr}).SetUpdate(bson.M{"$set": types.Address{
		Address:       addr,
		BalanceFloat:  balanceFloat,
		BalanceString: balanceString,
		IsContract:    isContract,
	}})
	return updateAddressOperation
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

	var addrs []*types.Address
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
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

//endregion Address
