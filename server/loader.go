// Package server
package server

import (
	"context"
	"math/big"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *infoServer) LoadBootData(ctx context.Context) error {
	lgr := s.logger
	lgr.Debug("Start load boot data")
	stats := s.dbClient.Stats(ctx)
	_ = s.cacheClient.SetTotalTxs(ctx, stats.TotalTransactions)
	_ = s.cacheClient.UpdateTotalHolders(ctx, stats.TotalAddresses, stats.TotalContracts)

	validators, err := s.kaiClient.Validators(ctx)
	if err != nil || len(validators) == 0 {
		lgr.Error("cannot get list validators", zap.Error(err))
		return err
	}

	if err := s.dbClient.ClearValidators(ctx); err != nil {
		return err
	}
	if err := s.dbClient.UpsertValidators(ctx, validators); err != nil {
		return err
	}
	for _, val := range validators {
		cfg.GenesisAddresses = append(cfg.GenesisAddresses, val.SmcAddress.String())
	}
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, cfg.TreasuryContractAddr)
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, cfg.StakingContractAddr)
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, cfg.KardiaDeployerAddr)
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, cfg.ParamsContractAddr)

	for _, addr := range cfg.GenesisAddresses {
		balance, _ := s.kaiClient.GetBalance(ctx, addr)
		balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
		balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
		addrInfo := &types.Address{
			Address:       addr,
			BalanceFloat:  balanceFloat,
			BalanceString: balance,
			IsContract:    false,
		}
		code, _ := s.kaiClient.GetCode(ctx, addr)
		if len(code) > 0 {
			addrInfo.IsContract = true
		}
		// write this address to db
		if err := s.dbClient.InsertAddress(ctx, addrInfo); err != nil {
			lgr.Debug("cannot insert address", zap.Error(err))
		}
	}
	return nil
}
