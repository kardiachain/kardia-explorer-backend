// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cContract = "Contracts"
	cABI      = "ContractABIs"
)

type ContractFilter struct {
}

type IContract interface {
	InsertContract(ctx context.Context, contract *types.Contract) error
	Contract(ctx context.Context, contractAddr string) (*types.Contract, error)
	UpdateContract(ctx context.Context, contract *types.Contract) error

	UpsertSMCABIByType(ctx context.Context, smcType, abi string) error
	SMCABIByType(ctx context.Context, smcType string) (string, error)
}

func (m *mongoDB) InsertContract(ctx context.Context, contract *types.Contract) error {
	if _, err := m.wrapper.C(cContract).Insert(contract); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) Contracts(ctx context.Context, filter ContractFilter) ([]*types.Contract, error) {
	var contracts []*types.Contract
	cursor, err := m.wrapper.C(cContract).Find(bson.M{})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (m *mongoDB) Contract(ctx context.Context, contractAddr string) (*types.Contract, error) {
	var contract *types.Contract
	err := m.wrapper.C(cContract).FindOne(bson.M{"address": contractAddr}).Decode(&contract)
	if err != nil {
		return nil, err
	}
	if contract.ABI == "" && contract.Type != "" {
		var smcABI *types.ContractABI
		err = m.wrapper.C(cABI).FindOne(bson.M{"type": contract.Type}).Decode(&smcABI)
		if err != nil {
			return nil, err
		}
		contract.ABI = smcABI.ABI
	}
	return contract, nil
}

func (m *mongoDB) UpdateContract(ctx context.Context, contract *types.Contract) error {
	return nil
}

func (m *mongoDB) UpsertSMCABIByType(ctx context.Context, smcType, abi string) error {
	upsertModel := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"type": smcType}).SetUpdate(bson.M{"$set": &types.ContractABI{
			Type: smcType,
			ABI:  abi,
		}}),
	}
	if _, err := m.wrapper.C(cABI).BulkWrite(upsertModel); err != nil {
		m.logger.Warn("cannot upsert new abi", zap.Error(err))
		return err
	}
	return nil
}

func (m *mongoDB) SMCABIByType(ctx context.Context, smcType string) (string, error) {
	var currABI *types.ContractABI
	if err := m.wrapper.C(cABI).FindOne(bson.M{"type": smcType}).Decode(&currABI); err != nil {
		return "", err
	}
	return currABI.ABI, nil
}
