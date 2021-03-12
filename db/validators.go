// Package db
package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cValidators = "Validators"
)

type IValidator interface {
	UpsertValidators(ctx context.Context, validators []*types.Validator) error
	Validators(ctx context.Context, filter ValidatorsFilter) ([]*types.Validator, error)
	ClearValidators(ctx context.Context) error

	UpsertValidator(ctx context.Context, validator *types.Validator) error
}

type ValidatorsFilter struct {
	Role int // [0:candidates, 1:validators, 2:proposer]
}

func (m *mongoDB) UpsertValidators(ctx context.Context, validators []*types.Validator) error {
	var models []mongo.WriteModel
	for _, v := range validators {
		models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"smcAddress": v.SmcAddress}).SetUpdate(bson.M{"$set": v}))
	}

	if _, err := m.wrapper.C(cValidators).BulkUpsert(models); err != nil {
		fmt.Println("Cannot write list model", err)
		return err
	}
	return nil
}

func (m *mongoDB) Validators(ctx context.Context, filter ValidatorsFilter) ([]*types.Validator, error) {
	var validators []*types.Validator
	cursor, err := m.wrapper.C(cValidators).Find(bson.M{})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &validators); err != nil {
		return nil, err
	}

	return validators, nil
}

func (m *mongoDB) ClearValidators(ctx context.Context) error {
	if _, err := m.wrapper.C(cValidators).RemoveAll(bson.M{}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertValidator(ctx context.Context, validator *types.Validator) error {
	var models []mongo.WriteModel
	models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"smcAddress": validator.SmcAddress}).SetUpdate(bson.M{"$set": validator}))

	if _, err := m.wrapper.C(cValidators).BulkUpsert(models); err != nil {
		fmt.Println("Cannot write list model", err)
		return err
	}
	return nil
}
