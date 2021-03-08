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
	cDelegator = "Delegators"
)

type IDelegators interface {
	InsertDelegators(ctx context.Context, delegators []*types.Delegator) error
	ClearDelegators(ctx context.Context, validatorSMCAddress string) error
}

func (m *mongoDB) InsertDelegators(ctx context.Context, delegators []*types.Delegator) error {
	lgr := m.logger.With(zap.String("method", "InsertDelegator"))
	var models []mongo.WriteModel
	for _, d := range delegators {
		models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"validatorSMC": d.ValidatorSMC, "address": d.Address}).SetUpdate(bson.M{"$set": d}))
	}

	if _, err := m.wrapper.C(cValidators).BulkWrite(models); err != nil {
		lgr.Warn("cannot write delegators", zap.Error(err))
		return err
	}
	return nil
}

func (m *mongoDB) ClearDelegators(ctx context.Context, validatorSMCAddr string) error {
	if _, err := m.wrapper.C(cDelegator).RemoveAll(bson.M{"validatorAddress": validatorSMCAddr}); err != nil {
		return err
	}

	return nil
}
