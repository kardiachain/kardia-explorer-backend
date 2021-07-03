package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/server/api"
	"go.uber.org/zap"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/utils"
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

	lgr, err := utils.NewLogger(serviceCfg)
	if err != nil {
		panic("cannot init logger")
	}
	lgr = lgr.With(zap.String("service_name", "receipts"))
	lgr.Info("Init receipts services", zap.Any("config", serviceCfg))
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

	kaiClientCfg := kardia.NewConfig(serviceCfg.KardiaPublicNodes, serviceCfg.KardiaTrustedNodes, lgr)
	kaiClient, err := kardia.NewKaiClient(kaiClientCfg)
	if err != nil {
		panic(err)
	}

	dbConfig := db.Config{
		DbAdapter: db.MGO,
		DbName:    serviceCfg.StorageDB,
		URL:       serviceCfg.StorageURI,
		Logger:    lgr,
		MinConn:   serviceCfg.StorageMinConn,
		MaxConn:   serviceCfg.StorageMaxConn,

		FlushDB: serviceCfg.StorageIsFlush,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		lgr.Error("cannot create db conn", zap.Error(err))
		panic(nil)
	}

	node, err := kClient.NewNode(serviceCfg.KardiaTrustedNodes[0], lgr)
	cacheCfg := cache.Config{
		Adapter:     cache.RedisAdapter,
		URL:         serviceCfg.CacheURL,
		DB:          serviceCfg.CacheDB,
		IsFlush:     serviceCfg.CacheIsFlush,
		BlockBuffer: serviceCfg.BufferedBlocks,
		Password:    serviceCfg.CachePassword,
		Logger:      lgr,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		lgr.Error("cannot create cache client", zap.Error(err))
		panic(err)
	}
	srv := new(api.Server).
		SetLogger(lgr).
		SetStorage(dbClient).
		SetCache(cacheClient).
		SetKaiClient(kaiClient).
		SetNode(node)

	if serviceCfg.IsReloadBootData {
		if err := srv.LoadBootContracts(ctx); err != nil {
			lgr.Panic("cannot load boot contracts", zap.Error(err))
		}
	}

	api.Start(srv, serviceCfg)
}
