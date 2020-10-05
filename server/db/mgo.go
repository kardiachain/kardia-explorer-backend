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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type mongoDB struct {
	logger  *zap.Logger
	wrapper *KaiMgo
	db      *mongo.Database
}

func newMongoDB(cfg ClientConfig) (*mongoDB, error) {
	ctx := context.Background()
	dbClient := &mongoDB{
		logger:  cfg.Logger,
		wrapper: &KaiMgo{},
	}
	mgoURI := fmt.Sprintf("mongodb://%s", cfg.URL)
	mgoClient, err := mongo.NewClient(options.Client().ApplyURI(mgoURI), options.Client().SetMinPoolSize(32), options.Client().SetMaxPoolSize(64))
	if err != nil {
		return nil, err
	}

	if err := mgoClient.Connect(context.Background()); err != nil {
		return nil, err
	}
	dbClient.wrapper.Database(mgoClient.Database(cfg.DbName))

	if cfg.FlushDB {
		dbClient.wrapper.DropDatabase(ctx)
	}

	return dbClient, nil
}

func (m *mongoDB) InsertBlock(ctx context.Context, block *types.Block) error {
	lgr := m.logger
	// todo @longnd: add block info ?
	// Upsert block into Blocks
	_, err := m.wrapper.C(cBlocks).Insert(block)
	if err != nil {
		lgr.Warn("cannot insert new block", zap.Error(err))
		return fmt.Errorf("cannot insert new block")
	}

	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": block.Height}); err != nil {
		return err
	}

	if err := m.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}

	return nil
}

// InsertTxs create bulk writer
func (m *mongoDB) InsertTxs(ctx context.Context, txs []*types.Transaction) error {
	var txsBulkWriter []mongo.WriteModel
	for _, tx := range txs {
		m.logger.Debug("Process tx", zap.String("tx", fmt.Sprintf("%#v", tx)))
		txModel := mongo.NewInsertOneModel().SetDocument(tx)
		txsBulkWriter = append(txsBulkWriter, txModel)
	}

	if _, err := m.wrapper.C("Transactions").BulkWrite(txsBulkWriter); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertTxs(ctx context.Context, txs []*types.Transaction) error {
	var txsBulkWriter []mongo.WriteModel
	for _, tx := range txs {
		m.logger.Debug("Process tx", zap.String("tx", fmt.Sprintf("%#v", tx)))
		txModel := mongo.NewInsertOneModel().SetDocument(tx)
		txsBulkWriter = append(txsBulkWriter, txModel)
	}

	if _, err := m.wrapper.C("Transactions").BulkWrite(txsBulkWriter); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertBlock(ctx context.Context, block *types.Block) error {
	panic("implement me")
}

func (m *mongoDB) IsBlockExist(ctx context.Context, block *types.Block) (bool, error) {
	panic("implement me")
}

func (m *mongoDB) BlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	panic("implement me")
}

func (m *mongoDB) ping() error {
	return nil
}

func (m *mongoDB) UpdateActiveAddress() error {
	panic("implement me")
}

func (m *mongoDB) dropCollection(collectionName string) {
	if _, err := m.wrapper.C(collectionName).RemoveAll(nil); err != nil {
		return
	}
}
