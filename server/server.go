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

package server

import (
	"github.com/kardiachain/explorer-backend/types"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
)

type Config struct {
	StorageAdapter db.Adapter
	StorageURI     string
	StorageDB      string
	StorageIsFlush bool
	MinConn        int
	MaxConn        int

	KardiaProtocol     kardia.Protocol
	KardiaURLs         []string
	KardiaTrustedNodes []string
	TotalValidators    int

	CacheAdapter cache.Adapter
	CacheURL     string
	CacheDB      int
	CacheIsFlush bool

	BlockBuffer int64

	HttpRequestSecret string

	VerifyBlockParam *types.VerifyBlockParam

	Metrics *metrics.Provider
	Logger  *zap.Logger
}

// Server instance kind of a router, which receive request from client (explorer)
// and control how we react those request
type Server struct {
	Logger *zap.Logger

	metrics *metrics.Provider

	VerifyBlockParam *types.VerifyBlockParam

	infoServer
}

func (s *Server) Metrics() *metrics.Provider { return s.metrics }

func New(cfg Config) (*Server, error) {
	cfg.Logger.Info("Create new server instance", zap.Any("config", cfg))
	dbConfig := db.Config{
		DbAdapter: cfg.StorageAdapter,
		DbName:    cfg.StorageDB,
		URL:       cfg.StorageURI,
		Logger:    cfg.Logger,
		MinConn:   cfg.MinConn,
		MaxConn:   cfg.MaxConn,

		FlushDB: cfg.StorageIsFlush,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		cfg.Logger.Debug("cannot create db client", zap.Error(err))
		return nil, err
	}

	kaiClientCfg := kardia.NewConfig(cfg.KardiaURLs, cfg.KardiaTrustedNodes, cfg.TotalValidators, cfg.Logger)
	kaiClient, err := kardia.NewKaiClient(kaiClientCfg)
	if err != nil {
		cfg.Logger.Debug("cannot create KaiClient", zap.Error(err))
		return nil, err
	}

	cacheCfg := cache.Config{
		Adapter:     cfg.CacheAdapter,
		URL:         cfg.CacheURL,
		DB:          cfg.CacheDB,
		IsFlush:     cfg.CacheIsFlush,
		BlockBuffer: cfg.BlockBuffer,
		Logger:      cfg.Logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return nil, err
	}
	avgMetrics := metrics.New()

	infoServer := infoServer{
		dbClient:          dbClient,
		cacheClient:       cacheClient,
		kaiClient:         kaiClient,
		HttpRequestSecret: cfg.HttpRequestSecret,
		verifyBlockParam:  cfg.VerifyBlockParam,
		logger:            cfg.Logger,
		metrics:           avgMetrics,
	}

	return &Server{
		Logger:     cfg.Logger,
		metrics:    avgMetrics,
		infoServer: infoServer,
	}, nil
}
