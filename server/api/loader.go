// Package api
package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/kardiachain/go-kardia/types/time"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.uber.org/zap"
)

func (s *Server) LoadBootContracts(ctx context.Context) error {
	// read and encode ABI base64
	krc20ABI, err := readAndEncodeABIFile("/abi/krc20.json")
	if err != nil {
		return err
	}
	krc721ABI, err := readAndEncodeABIFile("/abi/krc721.json")
	if err != nil {
		return err
	}
	paramsABI, err := readAndEncodeABIFile("/abi/params.json")
	if err != nil {
		return err
	}
	stakingABI, err := readAndEncodeABIFile("/abi/staking.json")
	if err != nil {
		return err
	}
	treasuryABI, err := readAndEncodeABIFile("/abi/treasury.json")
	if err != nil {
		return err
	}
	validatorABI, err := readAndEncodeABIFile("/abi/validator.json")
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
			Status:       types.ContractStatusVerified,
			IsVerified:   true,
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
	wd, _ := os.Getwd()
	abiFileContent, err := ioutil.ReadFile(path.Join(wd, filePath))
	if err != nil {
		return "", fmt.Errorf("cannot read ABI file %s", filePath)
	}
	return base64.StdEncoding.EncodeToString(abiFileContent), nil
}

func bootContracts() []*types.Contract {
	var bootSMCs []*types.Contract
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:       cfg.StakingContractName,
		Address:    cfg.StakingContractAddr,
		Bytecode:   cfg.StakingContractByteCode,
		CreatedAt:  time.Now().Unix(),
		Type:       "Staking",
		IsVerified: true,
	})
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:       cfg.ParamsContractName,
		Address:    cfg.ParamsContractAddr,
		Bytecode:   cfg.ParamsContractsByteCode,
		CreatedAt:  time.Now().Unix(),
		Type:       "Params",
		IsVerified: true,
	})
	bootSMCs = append(bootSMCs, &types.Contract{
		Name:       cfg.TreasuryContractName,
		Address:    cfg.TreasuryContractAddr,
		CreatedAt:  time.Now().Unix(),
		Type:       "Treasury",
		IsVerified: true,
	})

	return bootSMCs
}
