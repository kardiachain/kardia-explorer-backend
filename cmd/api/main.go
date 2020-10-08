package main

import (
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
)

func main() {
	serviceCfg, err := cfg.New()
	if err != nil {
		panic(err.Error())
		return
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic("cannot create logger")
	}

	srvCfg := server.Config{
		DBAdapter:       db.MGO,
		DBUrl:           serviceCfg.MongoURL,
		DBName:          serviceCfg.MongoDB,
		KardiaProtocol:  kardia.RPCProtocol,
		KardiaURL:       serviceCfg.KardiaURLs[0],
		CacheAdapter:    cache.RedisAdapter,
		CacheURL:        serviceCfg.CacheUrl,
		LockedAccount:   nil,
		IsFlushDatabase: true,
		Metrics:         nil,
		Logger:          logger,
	}
	srv, err := server.New(srvCfg)
	if err != nil {
		panic("cannot create server instance")
	}

	api.Start(srv)
}
