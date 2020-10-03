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
