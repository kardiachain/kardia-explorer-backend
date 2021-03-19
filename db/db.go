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

// DB define list API used by infoServer
type Client interface {
	IContract
	IValidator
	IDelegators
	Nodes
	IEvents
	IHolders
	IInternalTransaction
	ping() error
	dropCollection(collectionName string)
	dropDatabase(ctx context.Context) error

	// Stats
	UpdateStats(ctx context.Context, stats *types.Stats) error
	Stats(ctx context.Context) *types.Stats

	// Block details
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	IsBlockExist(ctx context.Context, blockHeight uint64) (bool, error)

	// Interact with blocks
	Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)
	InsertBlock(ctx context.Context, block *types.Block) error
	DeleteLatestBlock(ctx context.Context) (uint64, error)
	DeleteBlockByHeight(ctx context.Context, blockHeight uint64) error
	BlocksByProposer(ctx context.Context, proposer string, pagination *types.Pagination) ([]*types.Block, uint64, error)
	CountBlocksOfProposer(ctx context.Context, proposerAddress string) (int64, error)

	// Txs
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHeight(ctx context.Context, blockNumber uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	LatestTxs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)
	TxsCount(ctx context.Context) (uint64, error)

	// Tx detail
	TxByHash(ctx context.Context, txHash string) (*types.Transaction, error)

	// Interact with tx
	InsertTxs(ctx context.Context, txs []*types.Transaction) error
	InsertListTxByAddress(ctx context.Context, list []*types.TransactionByAddress) error

	// Address
	AddressByHash(ctx context.Context, addressHash string) (*types.Address, error)
	InsertAddress(ctx context.Context, address *types.Address) error
	UpdateAddresses(ctx context.Context, addresses []*types.Address) error
	GetTotalAddresses(ctx context.Context) (uint64, uint64, error)
	GetListAddresses(ctx context.Context, sortDirection int, pagination *types.Pagination) ([]*types.Address, error)
	Addresses(ctx context.Context) ([]*types.Address, error)

	AddressByName(ctx context.Context, name string) ([]*types.Address, error)
	ContractByName(ctx context.Context, name string) ([]*types.Contract, error)

	// Proposal
	AddVoteToProposal(ctx context.Context, proposalInfo *types.ProposalDetail, voteOption uint64) error
	UpsertProposal(ctx context.Context, proposalInfo *types.ProposalDetail) error
	ProposalInfo(ctx context.Context, proposalID uint64) (*types.ProposalDetail, error)
	GetListProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalDetail, uint64, error)
}

func NewClient(cfg Config) (Client, error) {
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid db config")
	}
}
