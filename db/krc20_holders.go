package db

import (
	"context"

	"github.com/kardiachain/go-kardia/lib/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var cKRC20Holders = "KRC20Holders"

type IKRC20Holder interface {
	createKRC20HoldersCollectionIndexes() []mongo.IndexModel
	UpsertKRC20Holders(ctx context.Context, holdersInfo []*types.KRC20Holder) error
	UpdateKRC20Holders(ctx context.Context, holdersInfo []*types.KRC20Holder) error
	KRC20Holders(ctx context.Context, filter *types.KRC20HolderFilter) ([]*types.KRC20Holder, uint64, error)
	RemoveKRC20Holder(ctx context.Context, holder *types.KRC20Holder) error

	RemoveKRC20Holders(ctx context.Context) error
}

func (m *mongoDB) createKRC20HoldersCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"balanceFloat": -1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"holderAddress": 1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) RemoveKRC20Holders(ctx context.Context) error {
	if _, err := m.wrapper.C(cKRC20Holders).RemoveAll(bson.M{"balance": "0"}); err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) UpsertKRC20Holders(ctx context.Context, holdersInfo []*types.KRC20Holder) error {
	holdersBulkWriter := make([]mongo.WriteModel, len(holdersInfo))
	for i := range holdersInfo {
		holdersInfo[i].HolderAddress = common.HexToAddress(holdersInfo[i].HolderAddress).String()
		txModel := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"holderAddress": holdersInfo[i].HolderAddress, "contractAddress": holdersInfo[i].ContractAddress}).SetUpdate(bson.M{"$set": holdersInfo[i]})
		holdersBulkWriter[i] = txModel
	}
	if len(holdersBulkWriter) > 0 {
		if _, err := m.wrapper.C(cKRC20Holders).BulkWrite(holdersBulkWriter); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) RemoveKRC20Holder(ctx context.Context, holder *types.KRC20Holder) error {
	if _, err := m.wrapper.C(cKRC20Holders).Remove(bson.M{"holderAddress": holder.HolderAddress, "contractAddress": holder.ContractAddress}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpdateKRC20Holders(ctx context.Context, holdersInfo []*types.KRC20Holder) error {
	holdersBulkWriter := make([]mongo.WriteModel, len(holdersInfo))
	for i := range holdersInfo {
		txModel := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"holderAddress": holdersInfo[i].HolderAddress, "contractAddress": holdersInfo[i].ContractAddress}).SetUpdate(bson.M{"$set": holdersInfo[i]})
		holdersBulkWriter[i] = txModel
	}
	if len(holdersBulkWriter) > 0 {
		if _, err := m.wrapper.C(cKRC20Holders).BulkWrite(holdersBulkWriter); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) KRC20Holders(ctx context.Context, filter *types.KRC20HolderFilter) ([]*types.KRC20Holder, uint64, error) {
	var (
		holders []*types.KRC20Holder
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
	var opts []*options.FindOptions
	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)), options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cKRC20Holders).Find(crit, opts...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()

	if err := cursor.All(ctx, &holders); err != nil {
		return nil, 0, err
	}

	total, err := m.wrapper.C(cKRC20Holders).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return holders, uint64(total), nil
}
