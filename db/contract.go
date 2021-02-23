// Package db
package db

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cContract = "Contracts"
)

type ContractFilter struct {
}

type IContract interface {
	InsertContract(ctx context.Context, contract *types.Contract) error
	InsertContracts(ctx context.Context, contracts []*types.Contract) error
	Contracts(ctx context.Context, filter ContractFilter) ([]*types.Contract, error)
	Contract(ctx context.Context, contractAddr string) (*types.Contract, error)
	UpdateContract(ctx context.Context, contract *types.Contract) error
}

func (m *mongoDB) InsertContract(ctx context.Context, contract *types.Contract) error {
	if _, err := m.wrapper.C(cContract).Insert(contract); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) InsertContracts(ctx context.Context, contracts []*types.Contract) error {
	for _, c := range contracts {
		//todo: handle error
		_ = m.InsertContract(ctx, c)
	}

	return nil
}

func (m *mongoDB) Contracts(ctx context.Context, filter ContractFilter) ([]*types.Contract, error) {
	return nil, nil
}

func (m *mongoDB) Contract(ctx context.Context, contractAddr string) (*types.Contract, error) {
	return nil, nil
}

func (m *mongoDB) UpdateContract(ctx context.Context, contract *types.Contract) error {
	return nil
}
