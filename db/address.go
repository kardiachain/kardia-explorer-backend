// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

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
