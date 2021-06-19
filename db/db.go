// Package db
package db

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

type Adapter string

const (
	MGO Adapter = "mgo"
)

type Config struct {
	DbAdapter Adapter
	DbName    string
	URL       string
	MinConn   int
	MaxConn   int
	FlushDB   bool

	Logger *zap.Logger
}

type Nodes interface {
	UpsertNode(ctx context.Context, node *types.NodeInfo) error
	Nodes(ctx context.Context) ([]*types.NodeInfo, error)
	RemoveNode(ctx context.Context, id string) error
}

type Client interface {
	ping() error
	dropCollection(collectionName string)
	dropDatabase(ctx context.Context) error

	IContract
	IValidator
	IDelegators
	Nodes
	IEvents
	IHolders
	IInternalTransaction
	IReceipt
	IBlock
	ITxs
	IAddress
	IProposal

	// Stats
	UpdateStats(ctx context.Context, stats *types.Stats) error
	Stats(ctx context.Context) *types.Stats
}

func NewClient(cfg Config) (Client, error) {
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid db config")
	}
}
