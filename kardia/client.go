/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
// Package kardia
package kardia

import (
	"context"

	"github.com/kardiachain/explorer-backend/types"
	"go.uber.org/zap"
)

type ClientInterface interface {
	LatestBlockNumber(ctx context.Context) (uint64, error)
	BlockByHash(ctx context.Context, hash string) (*types.Block, error)
	BlockByHeight(ctx context.Context, height uint64) (*types.Block, error)
	BlockHeaderByHash(ctx context.Context, hash string) (*types.Header, error)
	BlockHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error)
	GetTransaction(ctx context.Context, hash string) (*types.Transaction, bool, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error)
	BalanceAt(ctx context.Context, account string, blockHeightOrHash interface{}) (string, error)
	StorageAt(ctx context.Context, account string, key string, blockNumber uint64) ([]byte, error)
	CodeAt(ctx context.Context, account string, blockNumber uint64) ([]byte, error)
	NonceAt(ctx context.Context, account string) (uint64, error)
	SendRawTransaction(ctx context.Context, tx *types.Transaction) error
	Peers(ctx context.Context) (*types.PeerInfo, error)
	NodesInfo(ctx context.Context) ([]*types.NodeInfo, error)
	Datadir(ctx context.Context) (string, error)
	Validator(ctx context.Context, rpcURL string) (*types.Validator, error)
	Validators(ctx context.Context) ([]*types.Validator, error)
}

type Config struct {
	rpcURL            []string
	trustedNodeRPCURL []string
	lgr               *zap.Logger
}

func NewConfig(rpcURL []string, trustedNodeRPCURL []string, lgr *zap.Logger) *Config {
	return &Config{
		rpcURL:            rpcURL,
		trustedNodeRPCURL: trustedNodeRPCURL,
		lgr:               lgr,
	}
}
