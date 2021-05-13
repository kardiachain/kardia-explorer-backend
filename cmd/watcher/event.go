// Package main
package main

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/handler"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func runEventSubscriber(ctx context.Context, serviceCfg cfg.ExplorerConfig) {
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
		return
	}

	go h.ProcessNewEventLog(ctx)

	return
}
