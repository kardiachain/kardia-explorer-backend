// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cContract = "Contracts"
)

type ContractFilter struct {
}

type IContract interface {
	InsertContract(ctx context.Context, contract *types.Contract) error
	Contract(ctx context.Context, contractAddr string) (*types.Contract, error)
	UpdateContract(ctx context.Context, contract *types.Contract) error
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

	return contract, nil
}

func (m *mongoDB) UpdateContract(ctx context.Context, contract *types.Contract) error {
	return nil
}
