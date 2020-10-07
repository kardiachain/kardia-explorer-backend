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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	cBlocks = "Blocks"
	cTxs    = "Transactions"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"

)

type mongoDB struct {
	logger  *zap.Logger
	wrapper *KaiMgo
	db      *mongo.Database
}

func newMongoDB(cfg ClientConfig) (*mongoDB, error) {
	ctx := context.Background()
	dbClient := &mongoDB{
		logger:  cfg.Logger,
		wrapper: &KaiMgo{},
	}
	mgoURI := fmt.Sprintf("mongodb://%s", cfg.URL)
	mgoClient, err := mongo.NewClient(options.Client().ApplyURI(mgoURI), options.Client().SetMinPoolSize(32), options.Client().SetMaxPoolSize(64))
	if err != nil {
		return nil, err
	}

	if err := mgoClient.Connect(context.Background()); err != nil {
		return nil, err
	}
	dbClient.wrapper.Database(mgoClient.Database(cfg.DbName))

	if cfg.FlushDB {
		dbClient.wrapper.DropDatabase(ctx)
	}

	return dbClient, nil
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
		return err
	}

	if err := m.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}

	return nil
}

// InsertTxs create bulk writer
func (m *mongoDB) InsertTxs(ctx context.Context, txs []*types.Transaction) error {
	m.logger.Debug("Start insert txs", zap.Int("TxSize", len(txs)))
	var txsBulkWriter []mongo.WriteModel
	for _, tx := range txs {
		m.logger.Debug("Process tx", zap.String("tx", fmt.Sprintf("%#v", tx)))
		txModel := mongo.NewInsertOneModel().SetDocument(tx)
		txsBulkWriter = append(txsBulkWriter, txModel)
	}
	if len(txsBulkWriter) > 0 {
		if _, err := m.wrapper.C("Transactions").BulkWrite(txsBulkWriter); err != nil {
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

	if _, err := m.wrapper.C("Transactions").BulkWrite(txsBulkWriter); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertBlock(ctx context.Context, block *types.Block) error {
	return nil
}

func (m *mongoDB) IsBlockExist(ctx context.Context, block *types.Block) (bool, error) {
	panic("implement me")
}

func (m *mongoDB) BlockByHeight(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var c types.Block
	if err := m.wrapper.C("Blocks").FindOne(bson.M{"height": blockNumber}, options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0})).Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to get block: %v", err)
	}
	return &c, nil
}

func (m *mongoDB) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	var block types.Block
	err := m.wrapper.C("Blocks").FindOne(bson.M{"hash": blockHash}, options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0})).Decode(&block)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &block, nil
}

func (m *mongoDB) getLatestBlocks(ctx context.Context, filter *types.PaginationFilter) ([]*types.Block, error) {
	var blocks []*types.Block
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-height"),
		options.Find().SetProjection(bson.M{"height": 1, "time": 1, "validator": 1, "numTxs": 1}),
		options.Find().SetSkip(int64(filter.Skip)),
		options.Find().SetLimit(int64(filter.Limit)),
	}

	cursor, err := m.wrapper.C("Blocks").
		Find(bson.D{}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blocks: %v", err)
	}
	for cursor.Next(ctx) {
		block := &types.Block{}
		if err := cursor.Decode(&block); err != nil {
			return nil, err
		}
		m.logger.Debug("Get block success", zap.Any("block", block))
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (m *mongoDB) ping() error {
	return nil
}

func (m *mongoDB) UpdateActiveAddress() error {
	panic("implement me")
}

func (m *mongoDB) dropCollection(collectionName string) {
	if _, err := m.wrapper.C(collectionName).RemoveAll(nil); err != nil {
		return
	}
}

func (m *mongoDB) getActiveAddresses(fromDate time.Time) ([]*types.ActiveAddress, error) {
	var addresses []*types.ActiveAddress
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-updated_at"),
		options.Find().SetProjection(bson.M{"address": 1}),
	}

	cursor, err := m.wrapper.C("ActiveAddress").
		Find(bson.M{"updated_at": bson.M{"$gte": fromDate}}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blocks: %v", err)
	}
	for cursor.Next(context.TODO()) {
		address := &types.ActiveAddress{}
		if err := cursor.Decode(address); err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (m *mongoDB) isContract(address string) (bool, error) {
	var c types.Address
	err := m.wrapper.C("Addresses").FindOne(bson.M{"address": address}, options.FindOne().SetProjection(bson.M{"contract": 1})).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if contract: %v", err)
	}
	return c.Contract, nil
}

func (m *mongoDB) getAddressByHash(address string) (*types.Address, error) {
	var c types.Address
	err := m.wrapper.C("Addresses").FindOne(bson.M{"address": address}).Decode(&c)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get address: %v", err)
	}

	//lazy calculation for number of transactions
	var transactionCounter int64 = 0
	if m.useTransactionsByAddress() {
		transactionCounter, err = m.wrapper.C("TransactionsByAddress").Count(bson.M{"address": address})
		if err != nil {
			return nil, fmt.Errorf("failed to get txs from TransactionsByAddress: %v", err)
		}
	} else {
		transactionCounter, err = m.wrapper.C("Transactions").Count(bson.M{"$or": []bson.M{{"from": address}, {"to": address}}})
		if err != nil {
			return nil, fmt.Errorf("failed to get txs from Transactions: %v", err)
		}
	}
	c.NumberOfTransactions = int(transactionCounter)
	return &c, nil
}

func (m *mongoDB) getTxByAddressAndNonce(ctx context.Context, address string, nonce int64) (*types.Transaction, error) {
	var tx types.Transaction
	err := m.wrapper.C("Transactions").FindOne(bson.M{"from": address, "nonce": nonce}).Decode(&tx)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tx: %v", err)
	}
	return m.ensureReceipt(ctx, &tx)
}

func (m *mongoDB) getTxByHash(ctx context.Context, hash string) (*types.Transaction, error) {
	var tx types.Transaction
	err := m.wrapper.C("Transactions").FindOne(bson.M{"hash": hash}).Decode(&tx)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tx: %v", err)
	}
	return m.ensureReceipt(ctx, &tx)
}

// ensureReceipt does lazy loads receipt info if necessary.
func (m *mongoDB) ensureReceipt(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	panic("implement")
	//if tx.ReceiptReceived {
	//	return tx, nil
	//}
	//logger := m.logger.With(zap.String("tx", tx.TxHash))
	//receipt, err := m.kaiClient.TransactionReceipt(ctx, common.HexToHash(tx.TxHash))
	//if err != nil {
	//	logger.Warn("Failed to get transaction receipt", zap.Error(err))
	//} else {
	//	gasPrice, ok := new(big.Int).SetString(tx.GasPrice, 0)
	//	if !ok {
	//		logger.Error("Failed to parse gas price", zap.String("gasPrice", tx.GasPrice))
	//	}
	//	tmp := new(big.Int).Mul(gasPrice, big.NewInt(int64(receipt.GasUsed)))
	//	tx.GasFee = *tmp
	//	tx.ContractAddress = receipt.ContractAddress
	//	tx.Status = false
	//	if receipt.Status == 1 {
	//		tx.Status = true
	//	}
	//	tx.ReceiptReceived = true
	//	jsonValue, err := json.Marshal(receipt.Logs)
	//	if err != nil {
	//		logger.Error("Failed to marshal JSON receipt logs", zap.Error(err))
	//	}
	//	for _, l := range receipt.Logs {
	//		if err := m.UpdateActiveAddress(l.Address); err != nil {
	//			return nil, fmt.Errorf("failed to update active address: %s", err)
	//		}
	//	}
	//	tx.Logs = string(jsonValue)
	//	_, err = m.wrapper.C("Transactions").Upsert(bson.M{"hash": tx.TxHash}, tx)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to upsert tx: %v", err)
	//	}
	//}
	//return tx, nil
}

func (m *mongoDB) getTransactionList(ctx context.Context, address string, filter *types.TxsFilter) ([]*types.Transaction, error) {
	panic("implement")
}

func (m *mongoDB) getTokenHoldersList(contractAddress string, filter *types.PaginationFilter) ([]*types.TokenHolder, error) {
	var tokenHoldersList []*types.TokenHolder
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-balance_int"),
		options.Find().SetSkip(int64(filter.Skip)),
		options.Find().SetLimit(int64(filter.Limit)),
	}

	cursor, err := m.wrapper.C("TokensHolders").
		Find(bson.M{"contract_address": contractAddress}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get token holders list: %v", err)
	}

	for cursor.Next(context.TODO()) {
		tokenHolder := &types.TokenHolder{}
		if err := cursor.Decode(tokenHolder); err != nil {
			return nil, err
		}
		tokenHoldersList = append(tokenHoldersList, tokenHolder)
	}

	return tokenHoldersList, nil
}

func (m *mongoDB) getOwnedTokensList(ownerAddress string, filter *types.PaginationFilter) ([]*types.TokenHolder, error) {
	var tokenHoldersList []*types.TokenHolder
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-balance_int"),
		options.Find().SetSkip(int64(filter.Skip)),
		options.Find().SetLimit(int64(filter.Limit)),
	}
	cursor, err := m.wrapper.C("TokensHolders").
		Find(bson.M{"token_holder_address": ownerAddress}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned tokens list: %v", err)
	}
	for cursor.Next(context.TODO()) {
		tokenHolder := &types.TokenHolder{}
		cursor.Decode(tokenHolder)

		tokenHoldersList = append(tokenHoldersList, tokenHolder)
	}
	return tokenHoldersList, nil
}

// getInternalTokenTransfers gets token transfer events emitted by this contract.
func (m *mongoDB) getInternalTokenTransfers(contractAddress string, filter *types.InternalTxFilter) ([]*types.TokenTransfer, error) {
	var internalTransactionsList []*types.TokenTransfer
	query := bson.M{"contract_address": contractAddress}
	if filter.InternalAddress != "" {
		query = bson.M{"contract_address": contractAddress, "$or": []bson.M{{"from_address": filter.InternalAddress}, {"to_address": filter.InternalAddress}}}
	}

	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-block_number"),
		options.Find().SetSkip(int64(filter.PaginationFilter.Skip)),
		options.Find().SetLimit(int64(filter.PaginationFilter.Limit)),
	}

	cursor, err := m.wrapper.C("InternalTransactions").
		Find(query, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal txs list: %v", err)
	}
	for cursor.Next(context.TODO()) {
		tokenTransfer := &types.TokenTransfer{}
		if err := cursor.Decode(tokenTransfer); err != nil {
			return nil, err
		}
		internalTransactionsList = append(internalTransactionsList, tokenTransfer)
	}

	return internalTransactionsList, nil
}

// getHeldTokenTransfers gets token transfer events to or from this contract, for any token.
func (m *mongoDB) getHeldTokenTransfers(contractAddress string, filter *types.PaginationFilter) ([]*types.TokenTransfer, error) {
	var internalTransactionsList []*types.TokenTransfer
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-block_number"),
		options.Find().SetSkip(int64(filter.Skip)),
		options.Find().SetLimit(int64(filter.Limit)),
	}
	cursor, err := m.wrapper.C("InternalTransactions").
		Find(bson.M{"$or": []bson.M{{"from_address": contractAddress}, {"to_address": contractAddress}}}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal txs list: %v", err)
	}

	for cursor.Next(context.TODO()) {
		tokenTransfer := &types.TokenTransfer{}
		if err := cursor.Decode(tokenTransfer); err != nil {
			return nil, err
		}
		internalTransactionsList = append(internalTransactionsList, tokenTransfer)
	}
	return internalTransactionsList, nil
}

func (m *mongoDB) getContract(contractAddress string) (*types.Contract, error) {
	var contract *types.Contract
	err := m.wrapper.C("Contracts").FindOne(bson.M{"address": contractAddress}).Decode(&contract)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get contract: %v", err)
	}
	return contract, nil
}

func (m *mongoDB) getContractBlock(contractAddress string) (uint64, error) {
	var transaction *types.Transaction
	err := m.wrapper.C("Transactions").FindOne(bson.M{"contract_address": contractAddress}).Decode(&transaction)
	if err != nil {
		if err == mgo.ErrNotFound {
			return 0, errors.New("tx that deployed contract not found")
		}
		return 0, fmt.Errorf("failed to get tx that deployed contract: %v", err)
	}
	if transaction == nil {
		return 0, errors.New("tx that deployed contract not found")
	}
	return transaction.BlockNumber, nil
}

func (m *mongoDB) updateContract(contract *types.Contract) error {
	_, err := m.wrapper.C("Contracts").Upsert(bson.M{"address": contract.Address}, contract)
	if err != nil {
		return fmt.Errorf("failed to update contract: %v", err)
	}
	return nil
}

func (m *mongoDB) getContracts(filter *types.ContractsFilter) ([]*types.Address, error) {
	panic("implement")
	//var addresses []*types.Address
	//findQuery := bson.M{"contract": true}
	//if filter.TokenName != "" {
	//	findQuery["token_name"] = bson2.RegEx{regexp.QuoteMeta(filter.TokenName), "i"}
	//}
	//if filter.TokenSymbol != "" {
	//	findQuery["token_symbol"] = bson2.RegEx{regexp.QuoteMeta(filter.TokenSymbol), "i"}
	//}
	//if filter.ErcType != "" {
	//	findQuery["erc_types"] = filter.ErcType
	//}
	//if filter.SortBy == "" {
	//	filter.SortBy = "number_of_token_holders"
	//	filter.Asc = false
	//}
	//
	//contractQuery := bson.M{
	//	"attached_contract.valid": true,
	//}
	//if filter.ContractName != "" {
	//	contractQuery["attached_contract.contract_name"] = bson2.RegEx{regexp.QuoteMeta(filter.ContractName), "i"}
	//}
	//
	//sortDir := -1
	//if filter.Asc {
	//	sortDir = 1
	//}
	//sortQuery := bson.M{filter.SortBy: sortDir}
	//query := []bson.M{
	//	{"$match": findQuery},
	//	{"$lookup": bson.M{
	//		"from":         "Contracts",
	//		"localField":   "address",
	//		"foreignField": "address",
	//		"as":           "attached_contract",
	//	}},
	//	{"$match": contractQuery},
	//	{"$unwind": bson.M{
	//		"path": "$attached_contract",
	//	}},
	//	{"$sort": sortQuery},
	//	{"$skip": filter.Skip},
	//	{"$limit": filter.Limit},
	//}
	//err := m.mongo.
	//	C("Addresses").
	//	Pipe(query).
	//	All(&addresses)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to query contracts: %v", err)
	//}
	//return addresses, nil
}

func (m *mongoDB) getRichlist(filter *types.PaginationFilter, lockedAddresses []string) ([]*types.Address, error) {
	var addresses []*types.Address
	opts := []*options.FindOptions{
		m.wrapper.FindSetSort("-balance_float"),
		options.Find().SetSkip(int64(filter.Skip)),
		options.Find().SetLimit(int64(filter.Limit)),
	}

	cursor, err := m.wrapper.C("Addresses").Find(bson.M{"balance_float": bson.M{"$gt": 0}, "address": bson.M{"$nin": lockedAddresses}}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get rich list: %v", err)
	}

	for cursor.Next(context.TODO()) {
		address := &types.Address{}
		if err := cursor.Decode(address); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (m *mongoDB) updateStats() (*types.Stats, error) {
	numOfTotalTransactions, err := m.wrapper.C("Transactions").Count(nil)
	if err != nil {
		m.logger.Error("GetStats: Failed to get Total Transactions", zap.Error(err))
	}
	numOfLastWeekTransactions, err := m.wrapper.C("Transactions").Count(bson.M{"time": bson.M{"$gte": time.Now().AddDate(0, 0, -7)}})
	if err != nil {
		m.logger.Error("GetStats: Failed to get Last week Transactions", zap.Error(err))
	}
	numOfLastDayTransactions, err := m.wrapper.C("Transactions").Count(bson.M{"time": bson.M{"$gte": time.Now().AddDate(0, 0, -1)}})
	if err != nil {
		m.logger.Error("GetStats: Failed to get 24H Transactions", zap.Error(err))
	}
	stats := &types.Stats{
		NumberOfTotalTransactions:    numOfTotalTransactions,
		NumberOfLastWeekTransactions: numOfLastWeekTransactions,
		NumberOfLastDayTransactions:  numOfLastDayTransactions,
		UpdatedAt:                    time.Now(),
	}
	if _, err := m.wrapper.C("Stats").Insert(stats); err != nil {
		return nil, err
	}
	return stats, nil
}

func (m *mongoDB) getStats() (*types.Stats, error) {
	var s *types.Stats
	err := m.wrapper.C("Stats").FindOne(bson.D{}, m.wrapper.FindOneSetSort("-updated_at")).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %v", err)
	}
	return s, nil
}

func (m *mongoDB) getSignerStatsForRange(endTime time.Time, dur time.Duration) ([]types.SignerStats, error) {
	panic("implement")
	//var resp []bson.M
	//stats := []types.SignerStats{}
	//queryDayStats := []bson.M{bson.M{"$match": bson.M{"time": bson.M{"$gte": endTime.Add(dur)}}}, bson.M{"$group": bson.M{"_id": "$validator", "count": bson.M{"$sum": 1}}}}
	//err := m.mongo.C("Blocks").Pipe(queryDayStats).All(&resp)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to query signers stats: %v", err)
	//}
	//for _, el := range resp {
	//	addr := el["_id"].(string)
	//	if !common.IsHexAddress(addr) {
	//		return nil, fmt.Errorf("invalid hex address: %s", addr)
	//	}
	//	signerStats := types.SignerStats{SignerAddress: common.HexToAddress(addr), BlocksCount: el["count"].(int)}
	//	stats = append(stats, signerStats)
	//}
	//return stats, nil
}

func (m *mongoDB) getBlockRange(endTime time.Time, dur time.Duration) (types.BlockRange, error) {
	panic("implement")
	//var startBlock, endBlock types.Block
	//err := m.wrapper.C("Blocks").FindOne(bson.M{"time": bson.M{"$gte": endTime.Add(dur)}}, options.FindOne().SetProjection(bson.M{"height": 1}), m.wrapper.FindOneSetSort("time")).Decode(&startBlock)
	//if err != nil {
	//	return types.BlockRange{}, fmt.Errorf("failed to get start block number: %v", err)
	//}
	//err = m.wrapper.C("Blocks").FindOne(bson.M{"time": bson.M{"$gte": endTime.Add(dur)}}, options.FindOne().SetProjection(bson.M{"height": 1}), m.wrapper.FindOneSetSort("-time")).Decode(&endBlock)
	//if err != nil {
	//	return types.BlockRange{}, fmt.Errorf("failed to get end block number: %v", err)
	//}
	//return types.BlockRange{StartBlock: startBlock.Number, EndBlock: endBlock.Number}, nil
}

func (m *mongoDB) getSignersStats() ([]types.SignersStats, error) {
	var stats []types.SignersStats
	const day = -24 * time.Hour
	kvs := map[string]time.Duration{"daily": day, "weekly": 7 * day, "monthly": 30 * day}
	endTime := time.Now()
	for k, v := range kvs {
		blockRange, err := m.getBlockRange(endTime, v)
		if err != nil {
			return nil, fmt.Errorf("failed to get block range: %v", err)
		}
		signerStats, err := m.getSignerStatsForRange(endTime, v)
		if err != nil {
			return nil, fmt.Errorf("failed to get signer stats: %v", err)
		}
		stats = append(stats, types.SignersStats{BlockRange: blockRange, SignerStats: signerStats, Range: k})
	}
	return stats, nil
}

func (m *mongoDB) cleanUp() {
	//collectionNames, err := m.wrapper.CollectionNames()
	//if err != nil {
	//	m.logger.Error("Cannot get list of collections", zap.Error(err))
	//	return
	//}
	//for _, collectionName := range collectionNames {
	//	_, err := m.wrapper.C(collectionName).RemoveAll(nil)
	//	if err != nil {
	//		m.logger.Error("Failed to clean collection", zap.String("collection", collectionName), zap.Error(err))
	//		continue
	//	}
	//	m.logger.Info("Cleaned collection", zap.String("collection", collectionName))
	//}
}

func (m *mongoDB) useTransactionsByAddress() bool {
	//v, err := m.getDatabaseVersion()
	//if err != nil {
	//	m.logger.Error("Cannot get database version", zap.Error(err))
	//	return false
	//}
	//return v >= migrationTransactionsByAddress.ID
	return false
}
