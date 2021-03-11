// Package main
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func loadStakingBootData(ctx context.Context, cfg cfg.ExplorerConfig) error {
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

	validators, err := w.ValidatorsWithWorker(ctx)
	if err != nil {
		return err
	}
	if err := dbClient.UpsertValidators(ctx, validators); err != nil {
		return err
	}

	for _, v := range validators {
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

	return nil
}
