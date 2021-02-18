// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cNodes = "Nodes"
)

func (m *mongoDB) UpsertNode(ctx context.Context, node *types.NodeInfo) error {
	filter := bson.M{"id": node.ID}
	if _, err := m.wrapper.C(cNodes).Upsert(filter, node); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) Nodes(ctx context.Context) ([]*types.NodeInfo, error) {
	lgr := m.logger.With(zap.String("method", "Nodes"))
	filter := bson.M{}
	cursor, err := m.wrapper.C(cNodes).Find(filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			lgr.Warn("cannot close cursor", zap.Error(err))
			return
		}
	}()
	var nodes []*types.NodeInfo
	for cursor.Next(ctx) {
		var n *types.NodeInfo
		if err := cursor.Decode(&n); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func (m *mongoDB) RemoveNode(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	if _, err := m.wrapper.C(cNodes).RemoveAll(filter); err != nil {
		return err
	}
	return nil
}
