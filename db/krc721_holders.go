// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var cKRC721Holders = "KRC721Holders"

type IKRC721Holder interface {
	createKRC721HolderCollectionIndexes() []mongo.IndexModel
	UpsertKRC721Holders(ctx context.Context, holdersInfo []*types.KRC721Holder) error
	UpdateKRC721Holders(ctx context.Context, holdersInfo []*types.KRC721Holder) error
	KRC721Holders(ctx context.Context, filter types.KRC721HolderFilter) ([]*types.KRC721Holder, uint64, error)
	RemoveKRC721Holder(ctx context.Context, holder *types.KRC721Holder) error
}

func (m *mongoDB) createKRC721HolderCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"holderID": 1}, Options: options.Index().SetUnique(true).SetSparse(true)},
		{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"address": 1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) UpdateKRC721Holders(ctx context.Context, holdersInfo []*types.KRC721Holder) error {
	holdersBulkWriter := make([]mongo.WriteModel, len(holdersInfo))
	for i := range holdersInfo {
		holdersInfo[i].Address = common.HexToAddress(holdersInfo[i].Address).String()
		txModel := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"holderAddress": holdersInfo[i].Address, "contractAddress": holdersInfo[i].ContractAddress}).SetUpdate(bson.M{"$set": holdersInfo[i]})
		holdersBulkWriter[i] = txModel
	}
	if len(holdersBulkWriter) > 0 {
		if _, err := m.wrapper.C(cKRC721Holders).BulkWrite(holdersBulkWriter); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) KRC721Holders(ctx context.Context, filter types.KRC721HolderFilter) ([]*types.KRC721Holder, uint64, error) {
	var (
		holders []*types.KRC721Holder
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
	//opts := []*options.FindOptions{
	//	options.Find().SetHint(bson.M{"balanceFloat": -1}),
	//	options.Find().SetHint(bson.M{"contractAddress": 1}),
	//	options.Find().SetHint(bson.M{"holderAddress": 1}),
	//	options.Find().SetSort(bson.M{"balanceFloat": -1}),
	//}
	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)), options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cKRC721Holders).Find(crit, opts...)
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

	total, err := m.wrapper.C(cKRC721Holders).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return holders, uint64(total), nil
}

func (m *mongoDB) RemoveKRC721Holder(ctx context.Context, holder *types.KRC721Holder) error {
	if _, err := m.wrapper.C(cKRC721Holders).Remove(bson.M{"holderAddress": holder.Address, "contractAddress": holder.ContractAddress}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertKRC721Holders(ctx context.Context, holders []*types.KRC721Holder) error {
	holdersBulkWriter := make([]mongo.WriteModel, len(holders))
	for i := range holders {
		holders[i].HolderID = fmt.Sprintf("%s-%s", holders[i].ContractAddress, holders[i].TokenID)
		txModel := mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"holderID": holders[i].HolderID}).SetUpdate(bson.M{"$set": holders[i]})
		holdersBulkWriter[i] = txModel
	}
	if len(holdersBulkWriter) > 0 {
		if _, err := m.wrapper.C(cKRC721Holders).BulkWrite(holdersBulkWriter); err != nil {
			return err
		}
	}
	return nil
}
