// Package server
package server

import (
	"context"
	"time"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.uber.org/zap"
)

func (s *infoServer) ProcessActiveAddress(ctx context.Context, txs []*types.Transaction) error {
	lgr := s.logger.With(zap.String("method", "ProcessActiveAddress"))
	// update active addresses
	startTime := time.Now()
	s.filterContracts(ctx, txs)

	addrsMap := filterAddrSet(txs)
	getBalanceTime := time.Now()
	addrsList := s.getAddressBalances(ctx, addrsMap)
	lgr.Debug("GetAddressBalance time", zap.Duration("TotalTime", time.Since(getBalanceTime)))

	updateAddressTime := time.Now()
	if err := s.dbClient.UpdateAddresses(ctx, addrsList); err != nil {
		return err
	}
	lgr.Debug("UpdateAddressTime", zap.Duration("TotalTime", time.Since(updateAddressTime)))
	endTime := time.Since(startTime)
	s.metrics.RecordInsertActiveAddressTime(endTime)
	s.logger.Info("Total time for update addresses", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertActiveAddressTime()))
	startTime = time.Now()
	// todo: Recalculate new address later
	totalAddresses, err := s.dbClient.CountAddresses(ctx)
	if err == nil {
		if err := s.cacheClient.UpdateTotalAddresses(ctx, totalAddresses); err != nil {
			lgr.Error("cannot update total addresses", zap.Error(err))
		}
	}
	totalContracts, err := s.dbClient.CountContracts(ctx)
	if err == nil {
		if err := s.cacheClient.UpdateTotalContracts(ctx, totalContracts); err != nil {
			lgr.Error("cannot update total contracts", zap.Error(err))
		}
	}

	s.logger.Info("Total time for getting active addresses", zap.Duration("TimeConsumed", time.Since(startTime)))
	return nil
}

func (s *infoServer) filterContracts(ctx context.Context, txs []*types.Transaction) {
	lgr := s.logger
	for _, tx := range txs {
		if tx.ContractAddress != "" {

			c := &types.Contract{
				Address:      tx.ContractAddress,
				Bytecode:     tx.InputData,
				OwnerAddress: tx.From,
				TxHash:       tx.Hash,
				CreatedAt:    tx.Time.Unix(),
				Type:         cfg.SMCTypeNormal,              // Set normal by default
				Status:       types.ContractStatusUnverified, // Unverified by default
				IsVerified:   false,
			}
			lgr.Info("Detect new contract", zap.String("ContractAddress", c.Address), zap.String("TxHash", c.TxHash))
			if err := s.dbClient.InsertContract(ctx, c, nil); err != nil {
				lgr.Error("cannot insert new contract", zap.Error(err))
			}
		}
	}
}

func filterAddrSet(txs []*types.Transaction) map[string]*types.Address {
	addrs := make(map[string]*types.Address)
	for _, tx := range txs {
		addrs[tx.From] = &types.Address{
			Address:    tx.From,
			IsContract: false,
		}
		addrs[tx.To] = &types.Address{
			Address:    tx.To,
			IsContract: false,
		}
		addrs[tx.ContractAddress] = &types.Address{
			Address:    tx.ContractAddress,
			IsContract: true,
		}
	}
	delete(addrs, "")
	delete(addrs, "0x")
	return addrs
}
