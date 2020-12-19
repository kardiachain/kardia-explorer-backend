// Package server
package server

import (
	"context"

	"github.com/kardiachain/explorer-backend/types"
)

type Tx interface {
	Txs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)
	TotalTxs(ctx context.Context) (uint64, error)
	TxByHash(ctx context.Context, hash string) (*types.Transaction, error)
}

func (s *infoServer) Txs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error) {
	txs, err := s.cacheClient.LatestTransactions(ctx, pagination)
	if err != nil || txs == nil || len(txs) < pagination.Limit {
		txs, err = s.dbClient.LatestTxs(ctx, pagination)
		if err != nil {
			return nil, err
		}
	}
	return txs, nil
}

func (s *infoServer) TotalTxs(ctx context.Context) (uint64, error) {
	return s.cacheClient.TotalTxs(ctx), nil
}

func (s *infoServer) TxByHash(ctx context.Context, hash string) (*types.Transaction, error) {
	dbTx, err := s.dbClient.TxByHash(ctx, hash)
	if err == nil {
		return dbTx, nil
	}

	nTx, err := s.kaiClient.GetTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	receipt, err := s.kaiClient.GetTransactionReceipt(ctx, hash)
	if err != nil {
		//s.logger.Warn("cannot get receipt by hash from RPC:", zap.String("txHash", txHash))
	}
	if receipt != nil {
		nTx.Logs = receipt.Logs
		nTx.Root = receipt.Root
		nTx.Status = receipt.Status
		nTx.GasUsed = receipt.GasUsed
		nTx.ContractAddress = receipt.ContractAddress
	}

	return nTx, nil
}
