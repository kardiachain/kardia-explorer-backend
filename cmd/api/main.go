package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/cfg"
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

		KardiaURLs:         serviceCfg.KardiaURLs,
		KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,

		CacheAdapter:      cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:          serviceCfg.CacheURL,
		CacheDB:           serviceCfg.CacheDB,
		CacheIsFlush:      serviceCfg.CacheIsFlush,
		BlockBuffer:       serviceCfg.BufferedBlocks,
		HttpRequestSecret: serviceCfg.HttpRequestSecret,

		Metrics: nil,
		Logger:  logger,
	}
	srv, err := server.New(srvConfig)
	if err != nil {
		log.Panicf("cannot create server instance %s", err.Error())
	}

	e := echo.New()
	go func() {
		api.Start(e, srv, serviceCfg)
	}()
	if err := setupSentry(serviceCfg); err != nil {
		panic(err)
	}
	defer sentry.Flush(2 * time.Second)

	sentry.CaptureMessage("Test")

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	waitExit := make(chan bool)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
			cancel()
			if err := e.Shutdown(ctx); err != nil {
				panic(err)
			}
			waitExit <- true
		}
	}()
	<-waitExit
}

func setupSentry(cfg cfg.ExplorerConfig) error {
	opts := sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.ServerMode,
	}
	if err := sentry.Init(opts); err != nil {
		return err
	}
	return nil
}
