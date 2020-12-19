// Package server
package server

import (
	"context"

	"github.com/kardiachain/explorer-backend/types"
)

type Address interface {
	Balance(ctx context.Context, address string) (string, error)
	AddressTxs(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, error)
	TotalTxsOfAddress(ctx context.Context, address string) (uint64, error)
}

func (s *infoServer) Balance(ctx context.Context, address string) (string, error) {
	balance, err := s.kaiClient.GetBalance(ctx, address)
	if err != nil {
		return "0", err
	}
	return balance, nil
}

func (s *infoServer) AddressTxs(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, error) {
	txs, _, err := s.dbClient.TxsByAddress(ctx, address, pagination)
	if err != nil {
		return nil, err
	}
	return txs, nil

}

func (s *infoServer) TotalTxsOfAddress(ctx context.Context, address string) (uint64, error) {
	_, total, err := s.dbClient.TxsByAddress(ctx, address, &types.Pagination{
		Skip:  0,
		Limit: 10,
	})
	if err != nil {
		return 0, err
	}

	return total, nil
}
