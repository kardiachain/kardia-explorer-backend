package main

import (
	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/server/db"
)

func main() {
	serviceCfg, err := cfg.New()
	if err != nil {
		panic("cannot get configuration")
		return
	}

	srvCfg := server.Config{
		DBAdapter:       db.MGO,
		DBUrl:           serviceCfg.MongoURL,
		DBName:          serviceCfg.MongoDB,
		KardiaProtocol:  "",
		KardiaURL:       "",
		CacheAdapter:    "",
		CacheURL:        "",
		LockedAccount:   nil,
		Signers:         nil,
		IsFlushDatabase: false,
		Metrics:         nil,
		Logger:          nil,
	}
	srv, err := server.New(srvCfg)
	if err != nil {
		panic("cannot create server instance")
	}

	api.Start(srv)
}
