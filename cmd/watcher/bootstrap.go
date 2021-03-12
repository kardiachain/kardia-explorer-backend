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
		lgr.Debug("ValidatorInfo", zap.String("Name", v.Name))
		delegators, err := w.DelegatorsWithWorker(ctx, v.SmcAddress)
		if err != nil {
			lgr.Error("cannot load delegator", zap.String("validator", v.SmcAddress), zap.Error(err))
			return err
		}

		lgr.Info("delegator size", zap.Int("delegatorSize", len(delegators)))

		if err := dbClient.UpsertDelegators(ctx, delegators); err != nil {
			lgr.Error("cannot upsert delegators", zap.Error(err))
			return err
		}
	}
	totalValidator, err := dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return err
	}
	totalProposer, err := dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return err
	}
	totalCandidate, err := dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return err
	}
	totalUniqueDelegator, err := dbClient.UniqueDelegators(ctx)
	if err != nil {
		return err
	}

	totalValidatorStakedAmount, err := dbClient.GetStakedOfAddresses(ctx, validatorAddresses)
	if err != nil {
		return err
	}
	stakedAmountBigInt, ok := new(big.Int).SetString(totalValidatorStakedAmount, 10)
	if !ok {
		return fmt.Errorf("cannot load validator staked amount")
	}
	TotalDelegatorsStakedAmount := new(big.Int).Sub(totalStaked, stakedAmountBigInt)

	stats := &types.StakingStats{
		TotalValidators:            len(totalValidator),
		TotalProposers:             len(totalProposer),
		TotalCandidates:            len(totalCandidate),
		TotalDelegators:            totalUniqueDelegator,
		TotalStakedAmount:          totalStaked.String(),
		TotalValidatorStakedAmount: totalValidatorStakedAmount,
		TotalDelegatorStakedAmount: TotalDelegatorsStakedAmount.String(),
	}

	if err := cacheClient.UpdateStakingStats(ctx, stats); err != nil {
		return err
	}

	if err := dbClient.UpsertValidators(ctx, validators); err != nil {
		return err
	}
	lgr.Debug("Total time for boot ", zap.Any("time", time.Now().Sub(loadBootDataTime)))

	return nil
}
