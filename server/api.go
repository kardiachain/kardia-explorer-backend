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

package server

import (
	"context"
	"time"

	"github.com/kardiachain/go-kardiamain/lib/common"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	DefaultTimeout = 5 * time.Second
)

func (s *infoServer) LatestBlockNumber(ctx context.Context) (uint64, error) {
	toCtx, timeout := context.WithTimeout(ctx, DefaultTimeout)
	defer timeout()
	return s.kaiClient.LatestBlockNumber(toCtx)
}

func (s *infoServer) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	toCtx, timeout := context.WithTimeout(ctx, DefaultTimeout)
	defer timeout()
	return s.kaiClient.BlockByHash(toCtx, hash)
}
