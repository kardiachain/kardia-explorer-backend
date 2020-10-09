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
	"errors"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
)

type Config struct {
	DBAdapter db.Adapter
	DBUrl     string
	DBName    string

	KardiaProtocol kardia.Protocol
	KardiaURL      string

	CacheAdapter cache.Adapter
	CacheURL     string

	LockedAccount   []string
	IsFlushDatabase bool
	Metrics         *metrics.Provider
	Logger          *zap.Logger
}

// Server instance kind of a router, which receive request from client (explorer)
// and control how we react those request
type Server struct {
	Logger *zap.Logger

	metrics *metrics.Provider

	infoServer
}

func New(cfg Config) (*Server, error) {
	cfg.Logger.Info("Create new server instance", zap.Any("config", cfg))
	dbConfig := db.ClientConfig{
		DbAdapter: cfg.DBAdapter,
		DbName:    cfg.DBName,
		URL:       cfg.DBUrl,
		Logger:    cfg.Logger,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		return nil, err
	}

	urls := []string{"http://10.10.0.251:8551", "http://10.10.0.251:8548", "http://10.10.0.251:8549", "http://10.10.0.251:8550"}
	kaiClient, err := kardia.NewKaiClient(urls, cfg.Logger)
	if err != nil {
		return nil, errors.New("cannot create kai client")
	}

	cacheClient := cache.New(cache.Config{
		RedisUrl: cfg.CacheURL,
		Logger:   cfg.Logger,
	})

	infoServer := infoServer{
		dbClient:    dbClient,
		cacheClient: cacheClient,
		kaiClient:   kaiClient,
		logger:      cfg.Logger,
	}

	return &Server{
		Logger:     cfg.Logger,
		infoServer: infoServer,
	}, nil
}
