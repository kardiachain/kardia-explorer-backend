// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	cLog = "Logs"
)

type ILog interface {
}

func (m *mongoDB) InsertLogs(ctx context.Context, logs []*types.Log) error {
	lgr := m.logger.With(zap.String("method", "InsertLogs"))
	var models []mongo.WriteModel
	for _, l := range logs {
		models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"address": l.Address, "txHash": l.TxHash, "index": l.Index}).SetUpdate(bson.M{"$set": l}))
	}

	if _, err := m.wrapper.C(cLog).BulkWrite(models); err != nil {
		lgr.Error("cannot insert logs", zap.Error(err))
		return err
	}
	return nil
}

func (m *mongoDB) Logs() {

}

func (m *mongoDB) Log() {

}
