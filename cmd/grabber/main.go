/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
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

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/server"
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

		KardiaURLs:         serviceCfg.KardiaPublicNodes,
		KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,

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
		StorageIsFlush: false,

		KardiaURLs:         serviceCfg.KardiaPublicNodes,
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

	// todo: Temp remove verify worker
	// todo: Redis still store unverified block
	//srvConfigForVerifying := server.Config{
	//	StorageAdapter: db.Adapter(serviceCfg.StorageDriver),
	//	StorageURI:     serviceCfg.StorageURI,
	//	StorageDB:      serviceCfg.StorageDB,
	//	StorageIsFlush: false,
	//
	//	KardiaURLs:         serviceCfg.KardiaPublicNodes,
	//	KardiaTrustedNodes: serviceCfg.KardiaTrustedNodes,
	//
	//	CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
	//	CacheURL:     serviceCfg.CacheURL,
	//	CacheDB:      serviceCfg.CacheDB,
	//	CacheIsFlush: serviceCfg.CacheIsFlush,
	//	BlockBuffer:  serviceCfg.BufferedBlocks,
	//
	//	VerifyBlockParam: serviceCfg.VerifyBlockParam,
	//
	//	Metrics: nil,
	//	Logger:  logger.With(zap.String("service", "verifier")),
	//}
	//verifySrv, err := server.New(srvConfigForVerifying)
	//if err != nil {
	//	logger.Panic(err.Error())
	//}

	// Start listener in new go routine
	go listener(ctx, srv, serviceCfg.ListenerInterval)
	backfillCtx, _ := context.WithCancel(context.Background())
	go backfill(backfillCtx, backfillSrv, serviceCfg.BackfillInterval)
	//verifyCtx, _ := context.WithCancel(context.Background())
	//go verify(verifyCtx, verifySrv, serviceCfg.VerifierInterval)
	<-waitExit
	logger.Info("Stopped")
}
