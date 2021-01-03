// Package server
package server

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/kardiachain/go-kardia/types/time"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
)

var (
	tenPoweredBy5  = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
	tenPoweredBy18 = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

type ValidatorWatcher interface {
	SyncValidators(ctx context.Context) error
}

type watcher struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	lgr *zap.Logger
}

func NewValidatorWatcher(cfg Config) (ValidatorWatcher, error) {
	cfg.Logger.Info("Create new server instance", zap.Any("config", cfg))
	dbConfig := db.Config{
		DbAdapter: cfg.StorageAdapter,
		DbName:    cfg.StorageDB,
		URL:       cfg.StorageURI,
		Logger:    cfg.Logger,
		MinConn:   cfg.MinConn,
		MaxConn:   cfg.MaxConn,

		FlushDB: cfg.StorageIsFlush,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		return nil, err
	}

	kaiClientCfg := kardia.NewConfig(cfg.KardiaURLs, cfg.KardiaTrustedNodes, cfg.Logger)
	kaiClient, err := kardia.NewKaiClient(kaiClientCfg)
	if err != nil {
		return nil, err
	}

	cacheCfg := cache.Config{
		Adapter:     cfg.CacheAdapter,
		URL:         cfg.CacheURL,
		DB:          cfg.CacheDB,
		IsFlush:     cfg.CacheIsFlush,
		BlockBuffer: cfg.BlockBuffer,
		Logger:      cfg.Logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return nil, err
	}

	validatorSrv := watcher{
		dbClient:    dbClient,
		cacheClient: cacheClient,
		kaiClient:   kaiClient,
		lgr:         cfg.Logger,
	}
	return &validatorSrv, nil

}

func (s *watcher) calculateStats(ctx context.Context, validators []*types.Validator) (*types.StakingStats, error) {
	stats := types.StakingStats{}
	var (
		totalStakedAmount = big.NewInt(0)
		selfStakedAmount  = big.NewInt(0)
	)
	for _, v := range validators {
		fmt.Printf("Validator: %+v \n", v)
		role := GetRoleByStatus(v.Status)
		switch role {
		case Proposer:
			stats.TotalProposers++
			stats.TotalValidators++
		case Validator:
			stats.TotalValidators++
		case Candidate:
			stats.TotalCandidates++
		}

		stats.TotalDelegators += v.TotalDelegators
		for _, d := range v.Delegators {
			stakedAmount, ok := big.NewInt(0).SetString(d.StakedAmount, 10)
			if !ok {
				// todo: add notify here
			}
			if d.Address == v.Address {
				selfStakedAmount.Add(selfStakedAmount, stakedAmount)
			}
			totalStakedAmount.Add(totalStakedAmount, stakedAmount)
		}
	}

	delegateStakedAmount := big.NewInt(0).Sub(totalStakedAmount, selfStakedAmount)
	stats.TotalStakedAmount = totalStakedAmount.String()
	stats.TotalValidatorStakedAmount = selfStakedAmount.String()
	stats.TotalDelegatorStakedAmount = delegateStakedAmount.String()
	return &stats, nil
}

//SyncValidators fetch validators info from network and update to storage and cache
func (s *watcher) SyncValidators(ctx context.Context) error {
	s.lgr.Info("Sync validators", zap.Time("Timeline", time.Now()))
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return err
	}

	stats, err := s.calculateStats(ctx, validators.Validators)
	if err != nil {
		return err
	}

	fmt.Printf("Stats: %+v \n", stats)

	if err := s.dbClient.UpsertValidators(ctx, validators.Validators); err != nil {
		return err
	}

	if err := s.cacheClient.UpdateValidators(ctx, validators); err != nil {
		return err
	}

	// Update proposer in Addresses
	var proposers []*types.Address
	for _, v := range validators.Validators {
		balance, err := s.kaiClient.GetBalance(ctx, v.Address.Hex())
		if err != nil {
			balance = "0"
		}

		proposers = append(proposers, &types.Address{
			Address: v.Address.Hex(),
			Balance: balance,
		})
	}
	if err := s.dbClient.UpdateAddresses(ctx, proposers); err != nil {
		return err
	}
	return nil
}

func convertValidatorInfo(val *types.Validator, totalStakedAmount *big.Int, status int) (*types.Validator, error) {
	var (
		err  error
		zero = new(big.Int).SetInt64(0)
	)
	if val.CommissionRate, err = convertBigIntToPercentage(val.CommissionRate); err != nil {
		return nil, err
	}
	if val.MaxRate, err = convertBigIntToPercentage(val.MaxRate); err != nil {
		return nil, err
	}
	if val.MaxChangeRate, err = convertBigIntToPercentage(val.MaxChangeRate); err != nil {
		return nil, err
	}
	if totalStakedAmount != nil && totalStakedAmount.Cmp(zero) == 1 && status == 2 {
		if val.VotingPowerPercentage, err = calculateVotingPower(val.StakedAmount, totalStakedAmount); err != nil {
			return nil, err
		}
	} else {
		val.VotingPowerPercentage = "0"
	}
	val.SigningInfo.IndicatorRate = 100 - float64(val.SigningInfo.MissedBlockCounter)/100
	return val, nil
}

func convertBigIntToPercentage(raw string) (string, error) {
	input, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", errors.New("cannot parse")
	}
	tmp := new(big.Int).Mul(input, tenPoweredBy18)
	result := new(big.Int).Div(tmp, tenPoweredBy18).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-16]+"."+result[len(result)-16:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func calculateVotingPower(raw string, total *big.Int) (string, error) {
	valStakedAmount, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", errors.New("cannot parse ")
	}
	tmp := new(big.Int).Mul(valStakedAmount, tenPoweredBy5)
	result := new(big.Int).Div(tmp, total).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-3]+"."+result[len(result)-3:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

type Role string

var (
	Proposer  Role = "Proposer"
	Validator Role = "Validator"
	Candidate Role = "Candidate"
)

func GetRoleByStatus(status uint8) Role {
	switch status {
	case 2:
		return Proposer
	case 1:
		return Validator
	case 0:
		return Candidate
	}
	// todo: Notify something wrong here
	return Candidate
}
