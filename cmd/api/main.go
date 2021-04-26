package main

import (
	"context"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/api"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/server"
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

		KardiaURLs:         serviceCfg.KardiaPublicNodes,
		KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,

		CacheAdapter:      cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:          serviceCfg.CacheURL,
		CacheDB:           serviceCfg.CacheDB,
		CacheIsFlush:      serviceCfg.CacheIsFlush,
		BlockBuffer:       serviceCfg.BufferedBlocks,
		HttpRequestSecret: serviceCfg.HttpRequestSecret,

		Metrics: nil,
		Logger:  logger,

		AwsAccessKeyId:     serviceCfg.AwsAccessKeyId,
		AwsSecretAccessKey: serviceCfg.AwsSecretAccessKey,
		AwsSecretRegion:    serviceCfg.AwsSecretRegion,

		UploaderBucket:     serviceCfg.UploaderBucket,
		UploaderAcl:        serviceCfg.UploaderAcl,
		UploaderKey:        serviceCfg.UploaderKey,
		UploaderPathAvatar: serviceCfg.UploaderPathAvatar,
	}
	srv, err := server.New(srvConfig)
	if err != nil {
		logger.Panic("cannot create server instance %s", zap.Error(err))
	}
	ctx := context.Background()

	if serviceCfg.IsReloadBootData {
		//if err := srv.LoadBootData(ctx); err != nil {
		//	logger.Panic("cannot load boot data", zap.Error(err))
		//}
		if err := srv.LoadBootContracts(ctx); err != nil {
			logger.Panic("cannot load boot contracts", zap.Error(err))
		}
	}

	api.Start(srv, serviceCfg)
}
