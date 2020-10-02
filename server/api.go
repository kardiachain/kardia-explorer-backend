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

	"github.com/kardiachain/explorer-backend/kardia"
)

const (
	DefaultTimeout = 5 * time.Second
)

// apiServer handle how we retrieve and interact with other network for information
type apiServer struct {
	kaiClient kardia.Client
}

func (s *apiServer) latestBlockNumber(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	return s.kaiClient.LatestBlockNumber(ctx)
}
