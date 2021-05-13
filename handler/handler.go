// Package handler
package handler

import (
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
)

type Config struct {
	TrustedNodes []string
	PublicNodes  []string
	WSNodes      []string

	// DB config
	StorageAdapter db.Adapter
	StorageURI     string
	StorageDB      string

	CacheAdapter cache.Adapter
	CacheURL     string
	CacheDB      int

	Logger *zap.Logger
}

type Handler interface {
	IEvent
	IStakingHandler
}

type handler struct {
	// Internal
	w      *kardia.Wrapper
	db     db.Client
	cache  cache.Client
	logger *zap.Logger
}

func New(cfg Config) (Handler, error) {
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
