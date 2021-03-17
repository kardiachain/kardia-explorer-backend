// Package server
package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *infoServer) LoadBootData(ctx context.Context) error {
	lgr := s.logger
	lgr.Debug("Start load boot data")
	stats := s.dbClient.Stats(ctx)
	totalTxs, err := s.dbClient.TxsCount(ctx)
	if err != nil {
		s.logger.Warn("Cannot get total txs when boot", zap.Uint64("totalTxs", totalTxs), zap.Error(err))
	}
	if err = s.cacheClient.SetTotalTxs(ctx, totalTxs); err != nil {
		s.logger.Warn("Cannot set total txs to cache when boot", zap.Uint64("totalTxs", totalTxs), zap.Error(err))
	}
	if err = s.cacheClient.UpdateTotalHolders(ctx, stats.TotalAddresses, stats.TotalContracts); err != nil {
		s.logger.Warn("Cannot set total holders to cache when boot", zap.Uint64("totalAddresses", stats.TotalAddresses), zap.Uint64("totalContracts", stats.TotalContracts), zap.Error(err))
	}
	if err = s.dbClient.InsertAddress(ctx, &types.Address{
		Address:       "0x",
		BalanceString: "0",
		IsContract:    false,
	}); err != nil {
		s.logger.Warn("Cannot insert 0x address to db when boot", zap.Error(err))
	}

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
		cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
			Address: val.SmcAddress,
			Name:    val.Name,
		})
	}

	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.TreasuryContractAddr,
		Name:    cfg.TreasuryContractName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.StakingContractAddr,
		Name:    cfg.StakingContractName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.KardiaDeployerAddr,
		Name:    cfg.KardiaDeployerName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.ParamsContractAddr,
		Name:    cfg.ParamsContractName,
	})

	for i, addr := range cfg.GenesisAddresses {
		balance, _ := s.kaiClient.GetBalance(ctx, addr.Address)
		balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
		balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO

		cfg.GenesisAddresses[i].BalanceFloat = balanceFloat
		cfg.GenesisAddresses[i].BalanceString = balance
		code, _ := s.kaiClient.GetCode(ctx, addr.Address)
		if len(code) > 0 {
			cfg.GenesisAddresses[i].IsContract = true
		}

		// write this address to db
		if err := s.dbClient.InsertAddress(ctx, cfg.GenesisAddresses[i]); err != nil {
			lgr.Debug("cannot insert address", zap.Error(err))
		}

	}
	fmt.Println("Finished load boot data")
	return nil
}

func (s *infoServer) LoadBootContracts(ctx context.Context) error {
	// read and encode ABI base64
	krc20ABI, err := readAndEncodeABIFile("./abi/krc20.json")
	if err != nil {
		return err
	}
	krc721ABI, err := readAndEncodeABIFile("./abi/krc721.json")
	if err != nil {
		return err
	}
	paramsABI, err := readAndEncodeABIFile("./abi/params.json")
	if err != nil {
		return err
	}
	stakingABI, err := readAndEncodeABIFile("./abi/staking.json")
	if err != nil {
		return err
	}
	treasuryABI, err := readAndEncodeABIFile("./abi/treasury.json")
	if err != nil {
		return err
	}
	validatorABI, err := readAndEncodeABIFile("./abi/validator.json")
	if err != nil {
		return err
	}
	// store some types of SMC ABI to db
	smcABIByType := []*types.ContractABI{
		{
			Type: "Staking",
			ABI:  stakingABI,
		},
		{
			Type: cfg.SMCTypeValidator,
			ABI:  validatorABI,
		},
		{
			Type: "Params",
			ABI:  paramsABI,
		},
		{
			Type: "Treasury",
			ABI:  treasuryABI,
		},
		{
			Type: cfg.SMCTypeKRC20,
			ABI:  krc20ABI,
		},
		{
			Type: cfg.SMCTypeKRC721,
			ABI:  krc721ABI,
		},
	}
	for _, smcABI := range smcABIByType {
		err = s.dbClient.UpsertSMCABIByType(ctx, smcABI.Type, smcABI.ABI)
		if err != nil {
			s.logger.Warn("Cannot insert SMC ABI to DB", zap.Error(err))
		}
		err = s.cacheClient.UpdateSMCAbi(ctx, cfg.SMCTypePrefix+smcABI.Type, smcABI.ABI)
		if err != nil {
			s.logger.Warn("Cannot insert SMC ABI to cache", zap.Error(err))
		}
	}

	// insert boot smc and ABI(type) to db
	validators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil || len(validators) == 0 {
		s.logger.Error("cannot get list validators from db", zap.Error(err))
		return err
	}
	var bootSMCs []*types.Contract
	for _, val := range validators {
		bootSMCs = append(bootSMCs, &types.Contract{
			Name:         val.Name,
			Address:      val.SmcAddress,
			OwnerAddress: cfg.StakingContractAddr,
			CreatedAt:    time.Now().Unix(),
			Type:         cfg.SMCTypeValidator,
		})
	}
	bootSMCs = append(bootSMCs, bootContracts()...)
	for _, smc := range bootSMCs {
		err = s.dbClient.InsertContract(ctx, smc, nil)
		if err != nil {
			s.logger.Warn("Cannot insert SMC to db", zap.Error(err))
		}
	}

	return nil
}

func readAndEncodeABIFile(filePath string) (string, error) {
	_, filename, _, _ := runtime.Caller(1)
	abiFileContent, err := ioutil.ReadFile(path.Join(path.Dir(filename), filePath))
	if err != nil {
		return "", fmt.Errorf("cannot read ABI file %s", filePath)
	}
	return base64.StdEncoding.EncodeToString(abiFileContent), nil
}

func bootContracts() []*types.Contract {
	var bootSMCs []*types.Contract
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:      cfg.StakingContractName,
		Address:   cfg.StakingContractAddr,
		Bytecode:  cfg.StakingContractByteCode,
		CreatedAt: time.Now().Unix(),
		Type:      "Staking",
	})
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:      cfg.ParamsContractName,
		Address:   cfg.ParamsContractAddr,
		Bytecode:  cfg.ParamsContractsByteCode,
		CreatedAt: time.Now().Unix(),
		Type:      "Params",
	})
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:      cfg.TreasuryContractName,
		Address:   cfg.TreasuryContractAddr,
		CreatedAt: time.Now().Unix(),
		Type:      "Treasury",
	})

	return bootSMCs
}
