// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/kardiachain/explorer-backend/types"
)

var (
	cValidators = "Validators"
)

type IValidator interface {
	Validators()
	Validator()
	UpdateValidators()
	UpdateValidator()
}

type ValidatorsFilter struct {
}

func (m *mongoDB) Validators(ctx context.Context, filter ValidatorsFilter) ([]*types.Validator, error) {
	return nil, nil
}

func (m *mongoDB) Validator() {

}

func (m *mongoDB) UpdateValidator() {

}

func (m *mongoDB) UpsertValidators(ctx context.Context, validators []*types.Validator) error {

	if _, err := m.wrapper.C(cValidators).Update(bson.M{}, validators); err != nil {
		return err
	}
	return nil
}
