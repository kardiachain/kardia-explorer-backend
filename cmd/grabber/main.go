// Package main
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

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

	runtime.GOMAXPROCS(runtime.NumCPU())
	serviceCfg, err := cfg.New()
	if err != nil {
		panic(err.Error())
	}

	logger, err := newLogger(serviceCfg)
	if err != nil {
		panic("cannot init logger")
	}
	logger.Info("Start grabber...")

	defer func() {
		if err := recover(); err != nil {
			logger.Error("cannot recover")
		}
		if err := logger.Sync(); err != nil {
			logger.Error("cannot sync log")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	waitExit := make(chan bool)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
			cancel()
			waitExit <- true
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
		TotalValidators:    serviceCfg.TotalValidators,

		CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:     serviceCfg.CacheURL,
		CacheDB:      serviceCfg.CacheDB,
		CacheIsFlush: serviceCfg.CacheIsFlush,
		BlockBuffer:  serviceCfg.BufferedBlocks,

		Metrics: nil,
		Logger:  logger.With(zap.String("service", "listener")),
	}
	srv, err := server.New(srvConfig)
	if err != nil {
		logger.Panic(err.Error())
	}

	// Try to setup new srv instance since if we use same instance, maybe we will meet pool limit conn for mgo
	srvConfigForBackfill := server.Config{
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
		Logger:  logger.With(zap.String("service", "backfill")),
	}
	backfillSrv, err := server.New(srvConfigForBackfill)
	if err != nil {
		logger.Panic(err.Error())
	}

	srvConfigForVerifying := server.Config{
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

		VerifyBlockParam: serviceCfg.VerifyBlockParam,

		Metrics: nil,
		Logger:  logger.With(zap.String("service", "verifier")),
	}
	verifySrv, err := server.New(srvConfigForVerifying)
	if err != nil {
		logger.Panic(err.Error())
	}

	// Start listener in new go routine
	// todo @longnd: Running multi goroutine same time
	go listener(ctx, srv, serviceCfg.ListenerInterval)
	backfillCtx, _ := context.WithCancel(context.Background())
	go backfill(backfillCtx, backfillSrv, serviceCfg.BackfillInterval)
	verifyCtx, _ := context.WithCancel(context.Background())
	go verify(verifyCtx, verifySrv, serviceCfg.VerifierInterval)
	//updateAddresses(ctx, true, 0, srv)
	<-waitExit
	logger.Info("Grabber stopping")
	logger.Info("Stopped")
}
