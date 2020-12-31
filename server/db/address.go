// Package db
package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type FindAddressFilter struct {
	IsHasName bool
}

func (m *mongoDB) FindAddress(ctx context.Context, filter FindAddressFilter) ([]*types.Address, error) {
	var addrs []*types.Address
	opts := []*options.FindOptions{
		options.Find().SetProjection(bson.M{
			"address": 1,
			"name":    1,
		}),
	}
	hasNameFilter := bson.M{"name": bson.M{"$ne": ""}}

	cursor, err := m.wrapper.C(cAddresses).Find(hasNameFilter, opts...)
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

func (m *mongoDB) AddressByHash(ctx context.Context, address string) (*types.Address, error) {
	var c types.Address
	err := m.wrapper.C(cAddresses).FindOne(bson.M{"address": address}).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %v", err)
	}
	return &c, nil
}

func (m *mongoDB) InsertAddress(ctx context.Context, address *types.Address) error {
	address.CalculateOrder()
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
		info.CalculateOrder()
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
