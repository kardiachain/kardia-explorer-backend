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
)

const (
	DefaultTimeout = 5 * time.Second
)

func (s *infoServer) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return s.kaiClient.LatestBlockNumber(ctx)
}

func (s *infoServer) BlockCacheSize(ctx context.Context) (int64, error) {
	return s.cacheClient.BlocksSize(ctx)
}
