// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
)

func (m *mongoDB) LatestBlockHeight(ctx context.Context) (uint64, error) {
	latestBlock, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if err != nil || len(latestBlock) == 0 {
		return 0, err
	}
	return latestBlock[0].Height, nil
}

func (m *mongoDB) Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var blocks []*types.Block
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"height": -1}),
		options.Find().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.Find().SetSkip(int64(pagination.Skip)),
		options.Find().SetLimit(int64(pagination.Limit)),
	}

	cursor, err := m.wrapper.C(cBlocks).
		Find(bson.D{}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blocks: %v", err)
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	for cursor.Next(ctx) {
		block := &types.Block{}
		if err := cursor.Decode(&block); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (m *mongoDB) BlockByHeight(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var block types.Block
	if err := m.wrapper.C(cBlocks).FindOne(bson.M{"height": blockNumber},
		options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.FindOne().SetHint(bson.M{"height": -1})).Decode(&block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (m *mongoDB) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	var block types.Block
	err := m.wrapper.C(cBlocks).FindOne(bson.M{"hash": blockHash},
		options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0}),
		options.FindOne().SetHint(bson.M{"hash": 1})).Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (m *mongoDB) IsBlockExist(ctx context.Context, blockHeight uint64) (bool, error) {
	var dbBlock types.Block
	err := m.wrapper.C(cBlocks).FindOne(bson.M{"height": blockHeight}, options.FindOne().SetProjection(bson.M{"txs": 0, "receipts": 0})).Decode(&dbBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *mongoDB) InsertBlock(ctx context.Context, block *types.Block) error {
	logger := m.logger
	// Upsert block into Blocks
	_, err := m.wrapper.C(cBlocks).Insert(block)
	if err != nil {
		logger.Warn("cannot insert new block", zap.Error(err))
		return fmt.Errorf("cannot insert new block")
	}

	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": block.Height}); err != nil {
		logger.Warn("cannot remove old block txs", zap.Error(err))
		return err
	}

	return nil
}

func (m *mongoDB) DeleteLatestBlock(ctx context.Context) (uint64, error) {
	blocks, err := m.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 1,
	})
	if err != nil {
		m.logger.Warn("cannot get old latest block", zap.Error(err))
		return 0, err
	}
	if len(blocks) == 0 {
		m.logger.Warn("there isn't any block in database now, nothing to delete", zap.Error(err))
		return 0, nil
	}
	if err := m.DeleteBlockByHeight(ctx, blocks[0].Height); err != nil {
		m.logger.Warn("cannot remove old latest block", zap.Error(err), zap.Uint64("latest block height", blocks[0].Height))
		return 0, err
	}
	return blocks[0].Height, nil
}

func (m *mongoDB) DeleteBlockByHeight(ctx context.Context, blockHeight uint64) error {
	if _, err := m.wrapper.C(cBlocks).RemoveAll(bson.M{"height": blockHeight}); err != nil {
		m.logger.Warn("cannot remove old latest block", zap.Error(err), zap.Uint64("latest block height", blockHeight))
		return err
	}
	if _, err := m.wrapper.C(cTxs).RemoveAll(bson.M{"blockNumber": blockHeight}); err != nil {
		m.logger.Warn("cannot remove old latest block txs", zap.Error(err), zap.Uint64("latest block height", blockHeight))
		return err
	}
	return nil
}

func (m *mongoDB) BlocksByProposer(ctx context.Context, proposer string, pagination *types.Pagination) ([]*types.Block, uint64, error) {
	var blocks []*types.Block
	opts := []*options.FindOptions(nil)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.D{{Key: "proposerAddress", Value: 1}, {Key: "time", Value: -1}}),
			options.Find().SetSort(bson.M{"time": -1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}
	cursor, err := m.wrapper.C(cBlocks).
		Find(bson.M{"proposerAddress": proposer}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get txs for block: %v", err)
	}
	for cursor.Next(ctx) {
		block := &types.Block{}
		if err := cursor.Decode(block); err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, block)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cBlocks).Count(bson.M{"proposerAddress": proposer})
	if err != nil {
		return nil, 0, err
	}
	return blocks, uint64(total), nil
}

func (m *mongoDB) CountBlocksOfProposer(ctx context.Context, proposerAddress string) (int64, error) {
	total, err := m.wrapper.C(cBlocks).Count(bson.M{"proposerAddress": proposerAddress})
	if err != nil {
		return 0, err
	}
	return total, nil
}
