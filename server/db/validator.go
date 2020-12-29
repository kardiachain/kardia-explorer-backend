// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

var (
	cValidators = "Validators"
	cDashboard  = "Dashboard"
)

type ValidatorsFilter struct {
	IsAll       bool
	IsCandidate bool
	IsProposer  bool
}

func (m *mongoDB) UpsertValidators(ctx context.Context, validators []*types.Validator) error {
	for _, v := range validators {
		if _, err := m.wrapper.C(cValidators).Upsert(bson.M{"address": v.Address}, v); err != nil {
			return err
		}
	}

	return nil
}

func (m *mongoDB) FindValidators(ctx context.Context, filter ValidatorsFilter) ([]*types.Validator, error) {
	lgr := m.logger.With(zap.String("method", "FindValidators"))
	var validators []*types.Validator
	cursor, err := m.wrapper.C(cValidators).Find(bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			lgr.Warn("cannot close cursor", zap.Error(err))
			return
		}
	}()
	for cursor.Next(ctx) {
		var v types.Validator
		if err := cursor.Decode(&v); err != nil {
			return nil, err
		}
		validators = append(validators, &v)
	}

	return validators, nil
}
