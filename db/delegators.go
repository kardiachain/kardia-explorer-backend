// Package db
package db

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kardiachain/go-kardia/types/time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	cDelegator = "Delegators"
)

type IDelegators interface {
	Delegators(ctx context.Context, filter DelegatorFilter) ([]*types.Delegator, error)
	Delegator(ctx context.Context, delegatorAddress string) (*types.Delegator, error)

	UpsertDelegators(ctx context.Context, delegators []*types.Delegator) error
	ClearDelegators(ctx context.Context, validatorSMCAddress string) error
	UpsertDelegator(ctx context.Context, delegator *types.Delegator) error
	UniqueDelegators(ctx context.Context) (int, error)
	GetStakedOfAddresses(ctx context.Context, addresses []string) (string, error)
	CountDelegators(ctx context.Context, filter DelegatorFilter) (int64, error)
}

type DelegatorFilter struct {
	ValidatorSMCAddress string
	Skip                int64
	Limit               int64
}

func (m *mongoDB) Delegator(ctx context.Context, delegatorAddress string) (*types.Delegator, error) {
	return nil, nil
}

func (m *mongoDB) CountDelegators(ctx context.Context, filter DelegatorFilter) (int64, error) {
	var mgoFilter []bson.M
	if filter.ValidatorSMCAddress != "" {
		mgoFilter = append(mgoFilter, bson.M{"validatorSMCAddress": filter.ValidatorSMCAddress})
	}
	var (
		total int64
		err   error
	)
	if len(mgoFilter) == 0 {
		total, err = m.wrapper.C(cValidators).Count(bson.M{})
		if err != nil {
			return 0, err
		}
		return total, nil
	} else {
		total, err = m.wrapper.C(cValidators).Count(bson.M{"$and": mgoFilter})
	}
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (m *mongoDB) Delegators(ctx context.Context, filter DelegatorFilter) ([]*types.Delegator, error) {
	var delegators []*types.Delegator
	var opts []*options.FindOptions
	opts = append(opts, options.Find().SetSkip(filter.Skip), options.Find().SetLimit(filter.Limit))
	var mgoFilter []bson.M
	if filter.ValidatorSMCAddress != "" {
		mgoFilter = append(mgoFilter, bson.M{"validatorSMCAddress": filter.ValidatorSMCAddress})
	}
	var (
		cursor *mongo.Cursor
		err    error
	)
	if len(mgoFilter) == 0 {
		cursor, err = m.wrapper.C(cDelegator).Find(bson.M{}, opts...)
	} else {
		cursor, err = m.wrapper.C(cDelegator).Find(bson.M{"$and": mgoFilter}, opts...)
	}
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &delegators); err != nil {
		return nil, err
	}

	return delegators, nil
}

func (m *mongoDB) UpsertDelegator(ctx context.Context, delegator *types.Delegator) error {
	lgr := m.logger.With(zap.String("method", "UpsertDelegator"))
	var models []mongo.WriteModel
	models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"validatorSMCAddress": delegator.ValidatorSMCAddress, "address": delegator.Address}).SetUpdate(bson.M{"$set": delegator}))

	if _, err := m.wrapper.C(cDelegator).BulkUpsert(models); err != nil {
		lgr.Error("Cannot write list model", zap.Error(err))
		return err
	}
	return nil
}

func (m *mongoDB) UpsertDelegators(ctx context.Context, delegators []*types.Delegator) error {
	lgr := m.logger.With(zap.String("method", "InsertDelegator"))
	lgr.Debug("Start upsert delegators")
	startUpsertTime := time.Now()
	var models []mongo.WriteModel
	for _, d := range delegators {
		models = append(models, mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"validatorSMCAddress": d.ValidatorSMCAddress, "address": d.Address}).SetUpdate(bson.M{"$set": d}))
	}

	if _, err := m.wrapper.C(cDelegator).BulkUpsert(models); err != nil {
		lgr.Warn("cannot write delegators", zap.Error(err))
		return err
	}
	lgr.Debug("Finished upsert", zap.Duration("Total", time.Now().Sub(startUpsertTime)))
	return nil
}

func (m *mongoDB) ClearDelegators(ctx context.Context, validatorSMCAddr string) error {
	if _, err := m.wrapper.C(cDelegator).RemoveAll(bson.M{"validatorAddress": validatorSMCAddr}); err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) UniqueDelegators(ctx context.Context) (int, error) {
	total, err := m.wrapper.C(cDelegator).Count(bson.M{})
	if err != nil {
		return 0, err
	}

	fmt.Println("total row", total)
	data, err := m.wrapper.C(cDelegator).Distinct("address", bson.M{})
	if err != nil {
		return 0, err
	}

	fmt.Println("Data", len(data))
	return len(data), nil
}

func (m *mongoDB) GetStakedOfAddresses(ctx context.Context, addresses []string) (string, error) {
	cursor, err := m.wrapper.C(cDelegator).Find(bson.M{"address": bson.M{"$in": addresses}})
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)

	var delegators []*types.Delegator
	if err := cursor.All(ctx, &delegators); err != nil {
		return "", err
	}
	total := big.NewInt(0)
	for _, d := range delegators {
		stakedAmount, ok := new(big.Int).SetString(d.StakedAmount, 10)
		if !ok {
			return "", err
		}
		total = new(big.Int).Add(total, stakedAmount)
	}
	return total.String(), nil
}

func (m *mongoDB) RemoveDelegator(ctx context.Context, delegatorAddress, validatorSMCAddress string) error {
	return nil
}
