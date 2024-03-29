package main

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/handler"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func runStakingSubscriber(ctx context.Context, serviceCfg cfg.ExplorerConfig) error {

	logger, err := utils.NewLogger(serviceCfg)
	if err != nil {
		panic(err.Error())
	}

	handlerCfg := handler.Config{
		TrustedNodes: serviceCfg.KardiaTrustedNodes,
		PublicNodes:  serviceCfg.KardiaPublicNodes,
		WSNodes:      serviceCfg.KardiaWSNodes,

		StorageAdapter: db.Adapter(serviceCfg.StorageDriver),
		StorageURI:     serviceCfg.StorageURI,
		StorageDB:      serviceCfg.StorageDB,

		CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:     serviceCfg.CacheURL,
		CacheDB:      serviceCfg.CacheDB,

		Logger: logger,
	}
	h, err := handler.New(handlerCfg)
	if err != nil {
		panic(err.Error())
	}

	if serviceCfg.IsReloadStakingBootData {
		if err := loadStakingBootData(ctx, serviceCfg); err != nil {
			panic(err.Error())
		}
	}

	go h.ReloadValidators(ctx)

	go h.SubscribeStakingEvent(ctx)
	go h.SubscribeValidatorEvent(ctx)
	return nil
}
