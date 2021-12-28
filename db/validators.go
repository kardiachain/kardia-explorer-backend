// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kardia/lib/common"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"

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
	Validator(ctx context.Context, validatorAddress string) (*types.Validator, error)
	ClearValidators(ctx context.Context) error

	UpsertValidator(ctx context.Context, validator *types.Validator) error
	UpdateProposers(ctx context.Context, proposerSMCAddresses []string) error
	RemoveValidator(ctx context.Context, validatorSMCAddress string) error
}

type ValidatorsFilter struct {
	Role int // [1:candidates, 2:validators, 3:proposer]
	Skip int
}

func (m *mongoDB) UpsertValidators(ctx context.Context, validators []*types.Validator) error {
	lgr := m.logger.With(zap.String("method", "UpsertValidators"))
	var (
		models         []mongo.WriteModel
		addressModels  []mongo.WriteModel
		contractModels []mongo.WriteModel
	)
	for _, v := range validators {
		lgr.Info("Update model", zap.Any("ValInfo", v))
		models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"smcAddress": v.SmcAddress}).SetUpdate(bson.M{"$set": v}))
		contractInfo, addressInfo, err := m.Contract(ctx, v.SmcAddress)
		if err != nil {
			m.logger.Warn("Cannot get validator info from db", zap.Error(err), zap.Any("validatorInfo", v))
			contractInfo = &types.Contract{}
			addressInfo = &types.Address{}
		}
		addressInfo.Name = v.Name
		addressInfo.Address = v.SmcAddress
		contractInfo.Type = cfg.SMCTypeValidator
		contractInfo.Name = v.Name
		contractInfo.Address = v.SmcAddress

		addressModels = append(addressModels, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"address": addressInfo.Address}).SetUpdate(bson.M{"$set": addressInfo}))
		contractModels = append(contractModels, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"address": contractInfo.Address}).SetUpdate(bson.M{"$set": contractInfo}))
	}

	if _, err := m.wrapper.C(cValidators).BulkUpsert(models); err != nil {
		lgr.Error("Cannot write validator models", zap.Error(err))
		return err
	}
	if _, err := m.wrapper.C(cAddresses).BulkUpsert(addressModels); err != nil {
		lgr.Error("Cannot write address info models", zap.Error(err))
		return err
	}
	if _, err := m.wrapper.C(cContract).BulkUpsert(contractModels); err != nil {
		fmt.Println("Cannot write contract info models", err)
		return err
	}
	return nil
}

func (m *mongoDB) Validators(ctx context.Context, filter ValidatorsFilter) ([]*types.Validator, error) {
	var validators []*types.Validator

	var mgoFilter []bson.M
	//opts := []*options.FindOptions{
	//
	//}
	//if filter.Skip != 0 {
	//	options.Find().SetSkip(filter.Skip)
	//}
	if filter.Role != 0 {
		// Using role-1 since default go int == 0
		mgoFilter = append(mgoFilter, bson.M{"role": filter.Role - 1})
	}
	var (
		cursor *mongo.Cursor
		err    error
	)
	if len(mgoFilter) == 0 {
		cursor, err = m.wrapper.C(cValidators).Find(bson.M{})
	} else {
		cursor, err = m.wrapper.C(cValidators).Find(bson.M{"$and": mgoFilter})
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()

	if err := cursor.All(ctx, &validators); err != nil {
		return nil, err
	}

	return validators, nil
}

func (m *mongoDB) Validator(ctx context.Context, validatorAddress string) (*types.Validator, error) {
	// Force checksum validator address before make query
	// todo: better to force all address into lowercase before insert into db
	validatorAddress = common.HexToAddress(validatorAddress).String()
	var validator *types.Validator
	if err := m.wrapper.C(cValidators).FindOne(bson.M{"address": validatorAddress}).Decode(&validator); err != nil {
		return nil, err
	}
	return validator, nil
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

func (m *mongoDB) UpdateProposers(ctx context.Context, proposerAddresses []string) error {
	// Bind array
	if _, err := m.wrapper.C(cValidators).UpdateMany(bson.M{}, bson.M{"$set": bson.M{"status": 0, "role": 0}}); err != nil {
		return err
	}
	if _, err := m.wrapper.C(cValidators).UpdateMany(bson.M{"address": bson.M{"$in": proposerAddresses}}, bson.M{"$set": bson.M{"status": 2, "role": 2}}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) RemoveValidator(ctx context.Context, validatorSMCAddress string) error {
	if _, err := m.wrapper.C(cValidators).Remove(bson.M{"smcAddress": validatorSMCAddress}); err != nil {
		return err
	}
	return nil
}
