// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
)

type FindTxsFilter struct {
	ContractAddress string `json:"contractAddress" bson:"contractAddress,omitempty"`
	To              string `json:"to" bson:"to,omitempty"`
}

type ITxs interface {
	InsertTxs(ctx context.Context, txs []*types.Transaction) error

	LatestTxs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)
	TxsByAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHeight(ctx context.Context, blockNumber uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error)

	TxsCount(ctx context.Context) (uint64, error)
	TxByHash(ctx context.Context, txHash string) (*types.Transaction, error)
	FilterTxs(ctx context.Context, filter *types.TxsFilter) ([]*types.Transaction, uint64, error)

	FindContractCreationTxs(ctx context.Context) ([]*types.Transaction, error)
}

func (m *mongoDB) TxsCount(ctx context.Context) (uint64, error) {
	total, err := m.wrapper.C(cTxs).Count(bson.M{})
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

func (m *mongoDB) FindContractCreationTxs(ctx context.Context) ([]*types.Transaction, error) {
	var txs []*types.Transaction
	var opts []*options.FindOptions
	var mgoFilters []bson.M

	mgoFilters = append(mgoFilters, bson.M{"contractAddress": bson.M{"$ne": ""}})
	mgoFilters = append(mgoFilters, bson.M{"contractAddress": bson.M{"$ne": "0x"}})
	mgoFilters = append(mgoFilters, bson.M{"status": types.TransactionStatusSuccess})

	cursor, err := m.wrapper.C(cTxs).Find(bson.M{"$and": mgoFilters}, opts...)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &txs); err != nil {
		return nil, err
	}

	return txs, nil
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

func (m *mongoDB) FilterTxs(ctx context.Context, filter *types.TxsFilter) ([]*types.Transaction, uint64, error) {
	var (
		txs  []*types.Transaction
		crit = bson.M{}
	)
	critBytes, err := bson.Marshal(filter)
	if err != nil {
		m.logger.Warn("Cannot marshal txs filter criteria", zap.Error(err))
	}
	err = bson.Unmarshal(critBytes, &crit)
	if err != nil {
		m.logger.Warn("Cannot unmarshal txs filter criteria", zap.Error(err))
	}
	var opts []*options.FindOptions
	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)), options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cTxs).Find(crit, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err = cursor.All(ctx, &txs); err != nil {
		return nil, 0, err
	}

	total, err := m.wrapper.C(cTxs).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return txs, uint64(total), nil
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
