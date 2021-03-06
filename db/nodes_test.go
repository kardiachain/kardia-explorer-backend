// Package db
package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func TestMGO_UpsertNode(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)
	nodes := &types.NodeInfo{
		ProtocolVersion: types.ProtocolVersion{},
		ID:              "123454678",
		ListenAddr:      "3000",
		Network:         "1",
		Version:         "1",
		Moniker:         "1",
		Peers: map[string]*types.PeerInfo{
			"Peer1": {
				Duration: 0,
				Moniker:  "1",
				RemoteIP: "",
			},
			"Peer2": {
				Duration: 0,
				Moniker:  "2",
				RemoteIP: "",
			},
			"Peer3": {
				Duration: 0,
				Moniker:  "3",
				RemoteIP: "",
			},
		},
		Other: types.DefaultNodeInfoOther{},
	}
	err = mgo.UpsertNode(ctx, nodes)
	assert.Nil(t, err)
}

func TestMGO_Nodes(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)
	nodes, err := mgo.Nodes(ctx)
	assert.Nil(t, err)
	mgo.logger.Info("Nodes", zap.Any("nodes", nodes))
}
