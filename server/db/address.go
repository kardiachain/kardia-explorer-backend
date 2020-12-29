// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
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
