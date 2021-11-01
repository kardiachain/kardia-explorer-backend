// Package main
package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kardiachain/go-kardia/types/time"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func loadStakingBootData(ctx context.Context, cfg cfg.ExplorerConfig) error {
	loadBootDataTime := time.Now()
	if !cfg.IsReloadBootData {
		return nil
	}
	logger, err := utils.NewLogger(cfg)
	if err != nil {
		return err
	}
	lgr := logger.With(zap.String("method", "loadingStakingBootData"))
	wrapperCfg := kardia.WrapperConfig{
		TrustedNodes: cfg.KardiaTrustedNodes,
		PublicNodes:  cfg.KardiaPublicNodes,
		WSNodes:      cfg.KardiaWSNodes,
		Logger:       logger,
	}
	w, err := kardia.NewWrapper(wrapperCfg)
	if err != nil {
		return err
	}

	dbConfig := db.Config{
		DbAdapter: db.Adapter(cfg.StorageDriver),
		DbName:    cfg.StorageDB,
		URL:       cfg.StorageURI,
		Logger:    logger,
		MinConn:   1,
		MaxConn:   1,

		FlushDB: cfg.StorageIsFlush,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		return err
	}

	cacheCfg := cache.Config{
		Adapter: cache.RedisAdapter,
		URL:     cfg.CacheURL,
		DB:      cfg.CacheDB,
		IsFlush: cfg.CacheIsFlush,
		Logger:  logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return err
	}

	validators, err := w.ValidatorsWithWorker(ctx)
	if err != nil {
		return err
	}

	// Calculate voting power
	totalStaked, err := w.TrustedNode().TotalStakedAmount(ctx)
	if err != nil {
		return err
	}
	var validatorAddresses []string
	for id, v := range validators {
		validatorAddresses = append(validatorAddresses, v.Address)
		votingPower, err := utils.CalculateVotingPower(v.StakedAmount, totalStaked)
		if err != nil {
			return err
		}
		validators[id].VotingPowerPercentage = votingPower
		lgr.Info("ValidatorInfo", zap.String("Name", v.Name))
		delegators, err := w.DelegatorsWithWorker(ctx, v.SmcAddress)
		if err != nil {
			lgr.Error("cannot load delegator", zap.String("validator", v.SmcAddress), zap.Error(err))
			return err
		}

		lgr.Info("Delegator", zap.Int("Size", len(delegators)))

		if err := dbClient.UpsertDelegators(ctx, delegators); err != nil {
			lgr.Error("cannot upsert delegators", zap.Error(err))
			return err
		}
	}
	totalValidator, err := dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		lgr.Error("cannot calculate total validators", zap.Error(err))
		return err
	}
	// todo: Do not hard code here
	totalProposer, err := dbClient.Validators(ctx, db.ValidatorsFilter{Role: 3})
	if err != nil {
		lgr.Error("cannot calculate total proposers", zap.Error(err))
		return err
	}
	// todo: Fix hard code
	totalCandidate, err := dbClient.Validators(ctx, db.ValidatorsFilter{Role: 1})
	if err != nil {
		lgr.Error("cannot calculate total candidates", zap.Error(err))
		return err
	}
	totalUniqueDelegator, err := dbClient.UniqueDelegators(ctx)
	if err != nil {
		lgr.Error("cannot calculate total unique delegators", zap.Error(err))
		return err
	}

	totalValidatorStakedAmount, err := dbClient.GetStakedOfAddresses(ctx, validatorAddresses)
	if err != nil {
		lgr.Error("cannot calculate total validator staked amount", zap.Error(err))
		return err
	}
	stakedAmountBigInt, ok := new(big.Int).SetString(totalValidatorStakedAmount, 10)
	if !ok {
		return fmt.Errorf("cannot load validator staked amount")
	}
	TotalDelegatorsStakedAmount := new(big.Int).Sub(totalStaked, stakedAmountBigInt)

	stats := &types.StakingStats{
		TotalValidators:            len(totalValidator) - len(totalCandidate),
		TotalProposers:             len(totalProposer),
		TotalCandidates:            len(totalCandidate),
		TotalDelegators:            totalUniqueDelegator,
		TotalStakedAmount:          totalStaked.String(),
		TotalValidatorStakedAmount: totalValidatorStakedAmount,
		TotalDelegatorStakedAmount: TotalDelegatorsStakedAmount.String(),
	}

	if err := cacheClient.UpdateStakingStats(ctx, stats); err != nil {
		lgr.Error("cannot update staking stats", zap.Error(err))
		return err
	}
	// Clear data before upsert
	if err := dbClient.ClearValidators(ctx); err != nil {
		lgr.Error("cannot clear validators", zap.Error(err))
		return err
	}
	if err := dbClient.UpsertValidators(ctx, validators); err != nil {
		lgr.Error("cannot upsert validators", zap.Error(err))
		return err
	}
	lgr.Info("Finished loading boot data ", zap.Any("Total", time.Now().Sub(loadBootDataTime)))

	return nil
}
