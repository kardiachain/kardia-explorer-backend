package main

import (
	"github.com/joho/godotenv"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	serviceCfg, err := cfg.New()
	if err != nil {
		panic(err.Error())
	}

	logger, err := newLogger(serviceCfg)
	if err != nil {
		panic("cannot init logger")
	}
	logger.Info("Start API server...")

	defer func() {
		if err := recover(); err != nil {
			logger.Error("cannot recover")
		}
		if err := logger.Sync(); err != nil {
			logger.Error("cannot sync log")
		}
	}()

	srvConfig := server.Config{
		StorageAdapter: db.Adapter(serviceCfg.StorageDriver),
		StorageURI:     serviceCfg.StorageURI,
		StorageDB:      serviceCfg.StorageDB,
		StorageIsFlush: serviceCfg.StorageIsFlush,

		KardiaProtocol:     kardia.Protocol(serviceCfg.KardiaProtocol),
		KardiaURLs:         serviceCfg.KardiaURLs,
		KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,

		CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:     serviceCfg.CacheURL,
		CacheDB:      serviceCfg.CacheDB,
		CacheIsFlush: serviceCfg.CacheIsFlush,
		BlockBuffer:  serviceCfg.BufferedBlocks,

		Metrics: nil,
		Logger:  logger,
	}
	srv, err := server.New(srvConfig)
	if err != nil {
		panic("cannot create server instance")
	}

	api.Start(srv, serviceCfg)
}
