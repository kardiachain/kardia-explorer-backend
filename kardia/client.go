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

	"github.com/kardiachain/go-kardiamain/lib/common"

	"github.com/kardiachain/explorer-backend/types"
)

type Client interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number uint64) (*types.Block, error)
	LatestBlockNumber(ctx context.Context) (uint64, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number uint64) (*types.Header, error)
	TxByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	TxSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error)
	TotalTransactionOfBlock(ctx context.Context, blockHash common.Hash) (uint, error)
}
