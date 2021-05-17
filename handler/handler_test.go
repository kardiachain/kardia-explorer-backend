// Package handler
package handler

import (
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
)

func setupTestHandler() (*handler, error) {
	lgr, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	cfg := Config{
		TrustedNodes:   []string{"https://dev-1.kardiachain.io"},
		PublicNodes:    []string{"https://dev-1.kardiachain.io"},
		WSNodes:        []string{"wss://ws-dev.kardiachain.io"},
		StorageAdapter: "mgo",
		StorageURI:     "mongodb://54.179.162.247:27017",
		StorageDB:      "explorer",
		CacheAdapter:   "redis",
		CacheURL:       "54.179.162.247:6379",
		CacheDB:        0,
		Logger:         lgr,
	}

	wrapperCfg := kardia.WrapperConfig{
		TrustedNodes: cfg.TrustedNodes,
		PublicNodes:  cfg.PublicNodes,
		WSNodes:      cfg.WSNodes,
		Logger:       cfg.Logger,
	}
	kardiaWrapper, err := kardia.NewWrapper(wrapperCfg)
	if err != nil {
		return nil, err
	}

	dbConfig := db.Config{
		DbAdapter: cfg.StorageAdapter,
		DbName:    cfg.StorageDB,
		URL:       cfg.StorageURI,
		Logger:    cfg.Logger,
		MinConn:   1,
		MaxConn:   4,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		return nil, err
	}

	cacheCfg := cache.Config{
		Adapter: cfg.CacheAdapter,
		URL:     cfg.CacheURL,
		DB:      cfg.CacheDB,
		Logger:  cfg.Logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return nil, err
	}

	return &handler{
		w:      kardiaWrapper,
		logger: cfg.Logger,
		db:     dbClient,
		cache:  cacheClient,
	}, nil
}
