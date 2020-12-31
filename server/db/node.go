// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/kardiachain/explorer-backend/types"
)

var (
	cNode = "Nodes"
)

func (m *mongoDB) UpsertNodes(ctx context.Context, nodes []*types.NodeInfo) error {
	for _, n := range nodes {
		if _, err := m.wrapper.C(cNode).Upsert(bson.M{"id": n.ID}, n); err != nil {
			return err
		}
	}

	return nil
}

type FindNodesFilter struct {
}

func (m mongoDB) Find(ctx context.Context, filter FindNodesFilter, pagination *types.Pagination) ([]*types.NodeInfo, error) {
	return nil, nil
}
