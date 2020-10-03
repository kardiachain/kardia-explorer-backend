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
// Package db
package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	cBlocks = "Blocks"
	cTxs    = "Transactions"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"
	//CBlocks = "Blocks"

)

type MongoDB struct {
	logger  *zap.Logger
	wrapper *KaiMgo
}

func (m *MongoDB) ping() error {
	return nil
}

// importBlock handle follow task
// - Upsert block into `Blocks` collections/table
// - Remove any txs if block exist (if re-import)
func (m *MongoDB) importBlock(ctx context.Context, block *types.Block) error {
	lgr := m.logger
	// todo @longnd: add block info ?
	// Upsert block into Blocks
	_, err := m.wrapper.C(cBlocks).Upsert(bson.M{"height": block.Number}, block)
	if err != nil {
		lgr.Warn("cannot insert new block", zap.Error(err))
		return fmt.Errorf("cannot insert new block")
	}

	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": block.Number}); err != nil {
		return err
	}

	for _, tx := range block.Txs {
		fmt.Printf("Tx %+v \n", tx)
	}
	return nil
}

func (m *MongoDB) updateActiveAddress() error {
	panic("implement me")
}
