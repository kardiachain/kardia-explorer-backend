package db

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var cInternalTxs = "InternalTransactions"

type IInternalTransaction interface {
	createInternalTxsCollectionIndexes() []mongo.IndexModel
	RemoveInternalTxs(ctx context.Context, filter *types.InternalTxsFilter) error
	UpdateInternalTxs(ctx context.Context, internalTxs []*types.TokenTransfer) error
	GetListInternalTxs(ctx context.Context, filter *types.InternalTxsFilter) ([]*types.TokenTransfer, uint64, error)
}

func (m *mongoDB) createInternalTxsCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"from": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"to": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"txHash": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"time": -1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) UpdateInternalTxs(ctx context.Context, internalTxs []*types.TokenTransfer) error {
	iTxsBulkWriter := make([]mongo.WriteModel, len(internalTxs))
	for i := range internalTxs {
		iTxs := mongo.NewInsertOneModel().SetDocument(internalTxs[i])
		iTxsBulkWriter[i] = iTxs
	}
	if len(iTxsBulkWriter) > 0 {
		if _, err := m.wrapper.C(cInternalTxs).BulkWrite(iTxsBulkWriter); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) RemoveInternalTxs(ctx context.Context, filter *types.InternalTxsFilter) error {
	var crit bson.M
	critBytes, err := bson.Marshal(filter)
	if err != nil {
		m.logger.Warn("Cannot marshal txs filter criteria", zap.Error(err))
	}
	err = bson.Unmarshal(critBytes, &crit)
	if err != nil {
		m.logger.Warn("Cannot unmarshal txs filter criteria", zap.Error(err))
	}
	if _, err = m.wrapper.C(cInternalTxs).RemoveAll(crit); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) GetListInternalTxs(ctx context.Context, filter *types.InternalTxsFilter) ([]*types.TokenTransfer, uint64, error) {
	opts := []*options.FindOptions{
		options.Find().SetSort(bson.M{"time": -1}),
	}
	var (
		iTxs    []*types.TokenTransfer
		andCrit []bson.M
	)
	if filter.Address != "" {
		orCrit := []bson.M{
			{"from": filter.Address},
			{"to": filter.Address},
		}
		andCrit = append(andCrit, bson.M{"$or": orCrit})
		//opts = append(opts, options.Find().SetHint(bson.M{"from": 1}))
		//opts = append(opts, options.Find().SetHint(bson.M{"to": 1}))
	}
	if filter.Contract != "" {
		andCrit = append(andCrit, bson.M{"contractAddress": filter.Contract})
		opts = append(opts, options.Find().SetHint(bson.M{"contractAddress": 1}))
	}
	if filter.TransactionHash != "" {
		andCrit = append(andCrit, bson.M{"txHash": filter.TransactionHash})
		opts = append(opts, options.Find().SetHint(bson.M{"txHash": 1}))
	}
	crit := bson.M{"$and": andCrit}

	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts /*options.Find().SetSkip(int64(filter.Pagination.Skip)),*/, options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cInternalTxs).Find(crit, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err := cursor.All(ctx, &iTxs); err != nil {
		return nil, 0, err
	}

	total, err := m.wrapper.C(cInternalTxs).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return iTxs, uint64(total), nil
}
