// Package db
package db

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *mongoDB) AllOldHolders(ctx context.Context) ([]*types.KRC20Holder, error) {
	cursor, err := m.wrapper.C("Holders").Find(bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	var holders []*types.KRC20Holder
	if err := cursor.All(ctx, &holders); err != nil {
		return nil, err
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return holders, nil
}
