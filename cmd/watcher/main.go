package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cache"
	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/db"
	"github.com/kardiachain/explorer-backend/server"
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
	logger.Info("Setup watcher...")

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

	var srvConfig = server.Config{
		StorageAdapter: db.Adapter(serviceCfg.StorageDriver),
		StorageURI:     serviceCfg.StorageURI,
		StorageDB:      serviceCfg.StorageDB,
		StorageIsFlush: serviceCfg.StorageIsFlush,

		KardiaURLs:         serviceCfg.KardiaURLs,
		KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,

		CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:     serviceCfg.CacheURL,
		CacheDB:      serviceCfg.CacheDB,
		CacheIsFlush: serviceCfg.CacheIsFlush,
		BlockBuffer:  serviceCfg.BufferedBlocks,

		Metrics: nil,
		Logger:  logger.With(zap.String("service", "listener")),
	}
	validatorWatcher, err := server.NewValidatorWatcher(srvConfig)
	if err != nil {
		logger.Panic(err.Error())
	}

	// Start watcher in new go routine
	go watchValidators(ctx, validatorWatcher, cfg.UpdateStatsInterval)
	<-waitExit
	logger.Info("Stopped")
}
