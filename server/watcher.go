// Package server
package server

import (
	"context"

	"github.com/kardiachain/go-kardia/types/time"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
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

//SyncValidators fetch validators info from network and update to storage and cache
func (s *watcher) SyncValidators(ctx context.Context) error {
	s.lgr.Info("Sync validators", zap.Time("Timeline", time.Now()))
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return err
	}

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
