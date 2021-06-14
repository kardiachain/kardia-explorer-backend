package db

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	cReceipts = "Receipts"
)

type IReceipt interface {
	InsertReceipts(ctx context.Context, receipts []*types.Receipt) error
	RemoveReceipt(ctx context.Context, receipt *types.Receipt) error
}

func (m *mongoDB) InsertReceipts(ctx context.Context, receipts []*types.Receipt) error {
	var (
		receiptsBulkWriter []mongo.WriteModel
	)
	for _, r := range receipts {
		txModel := mongo.NewInsertOneModel().SetDocument(r)
		receiptsBulkWriter = append(receiptsBulkWriter, txModel)
	}
	if len(receiptsBulkWriter) > 0 {
		if _, err := m.wrapper.C(cReceipts).BulkWrite(receiptsBulkWriter); err != nil {
			return err
		}
	}

	return nil
}

func (m *mongoDB) RemoveReceipt(ctx context.Context, receipt *types.Receipt) error {
	return nil
}
