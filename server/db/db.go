// Package db
package db

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Adapter string

const (
	MGO Adapter = "mongo"
)

type ClientConfig struct {
	DbAdapter Adapter
	DbName    string
	URL       string
	FlushDB   bool

	Logger *zap.Logger
}

// DB define list API used by infoServer
type Client interface {
	ping() error
	dropCollection(collectionName string)

	// Blocks
	Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)

	// Block details
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	IsBlockExist(ctx context.Context, block *types.Block) (bool, error)

	// Interact with block
	InsertBlock(ctx context.Context, block *types.Block) error
	UpsertBlock(ctx context.Context, block *types.Block) error

	// Txs
	Txs(ctx context.Context, pagination *types.Pagination)
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, error)
	TxsByBlockHeight(ctx context.Context, blockNumber uint64, pagination *types.Pagination) ([]*types.Transaction, error)
	TxsByAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, error)

	// Tx detail
	TxByHash(ctx context.Context, txHash string) (*types.Transaction, error)
	TxByNonce(ctx context.Context, nonce int64) (*types.Transaction, error)

	// Interact with tx
	InsertTxs(ctx context.Context, txs []*types.Transaction) error
	UpsertTxs(ctx context.Context, txs []*types.Transaction) error
	InsertTxByAddress(ctx context.Context, address, txHash string, createdAt int64) error

	// Interact with receipts
	InsertReceipts(ctx context.Context, block *types.Block) error
	UpsertReceipts(ctx context.Context, block *types.Block) error

	// Token
	TokenHolders(ctx context.Context, tokenAddress string, pagination *types.Pagination) ([]*types.TokenHolder, error)
	//InternalTxs(ctx context.Context)

	// Address
	AddressByHash(ctx context.Context, addressHash string) (*types.Address, error)
	OwnedTokensOfAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.TokenHolder, error)

	UpdateActiveAddresses(ctx context.Context, addresses []string) error
}

func NewClient(cfg ClientConfig) (Client, error) {
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid config")
	}
}
