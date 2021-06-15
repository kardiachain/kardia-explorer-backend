package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const cHolders = "Holders"

type IHolders interface {
	createHoldersCollectionIndexes() []mongo.IndexModel
	UpdateHolders(ctx context.Context, holdersInfo []*types.TokenHolder) error
	GetListHolders(ctx context.Context, filter *types.HolderFilter) ([]*types.TokenHolder, uint64, error)
}

func (m *mongoDB) createHoldersCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"balanceFloat": -1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"holderAddress": 1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) UpdateHolders(ctx context.Context, holdersInfo []*types.TokenHolder) error {
	holdersBulkWriter := make([]mongo.WriteModel, len(holdersInfo))
	for i := range holdersInfo {
		txModel := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"holderAddress": holdersInfo[i].HolderAddress, "contractAddress": holdersInfo[i].ContractAddress}).SetUpdate(bson.M{"$set": holdersInfo[i]})
		holdersBulkWriter[i] = txModel
	}
	if len(holdersBulkWriter) > 0 {
		if _, err := m.wrapper.C(cHolders).BulkWrite(holdersBulkWriter); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) GetListHolders(ctx context.Context, filter *types.HolderFilter) ([]*types.TokenHolder, uint64, error) {
	var (
		holders []*types.TokenHolder
		crit    = bson.M{}
	)
	critBytes, err := bson.Marshal(filter)
	if err != nil {
		m.logger.Warn("Cannot marshal holder filter criteria", zap.Error(err))
	}
	err = bson.Unmarshal(critBytes, &crit)
	if err != nil {
		m.logger.Warn("Cannot unmarshal holder filter criteria", zap.Error(err))
	}
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"balanceFloat": -1}),
		options.Find().SetHint(bson.M{"contractAddress": 1}),
		options.Find().SetHint(bson.M{"holderAddress": 1}),
		options.Find().SetSort(bson.M{"balanceFloat": -1}),
	}
	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)), options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cHolders).Find(crit, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err := cursor.All(ctx, &holders); err != nil {
		return nil, 0, err
	}

	total, err := m.wrapper.C(cHolders).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return holders, uint64(total), nil
}
