// Package server
package server

import (
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/metrics"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func createDevSrv() (*Server, error) {
	cfg := Config{
		StorageAdapter:     "mgo",
		StorageURI:         "mongodb://kardia.ddns.net:27017",
		StorageDB:          "explorerTest",
		StorageIsFlush:     true,
		MinConn:            4,
		MaxConn:            16,
		KardiaURLs:         []string{"https://dev-1.kardiachain.io"},
		KardiaTrustedNodes: []string{"https://dev-1.kardiachain.io"},
		CacheAdapter:       "redis",
		CacheURL:           "kardia.ddns.net:6379",
		CacheDB:            0,
		CacheIsFlush:       true,
		BlockBuffer:        20,
		HttpRequestSecret:  "httpTestSecret",
		VerifyBlockParam: &types.VerifyBlockParam{
			VerifyTxCount:   false,
			VerifyBlockHash: false,
		},
		Metrics: metrics.New(),
		Logger:  zap.L(),
	}

	return New(cfg)
}

func createTestSrv() (*Server, error) {
	cfg := Config{
		StorageAdapter:     "",
		StorageURI:         "",
		StorageDB:          "",
		StorageIsFlush:     false,
		MinConn:            0,
		MaxConn:            0,
		KardiaURLs:         nil,
		KardiaTrustedNodes: nil,
		CacheAdapter:       "",
		CacheURL:           "",
		CacheDB:            0,
		CacheIsFlush:       false,
		BlockBuffer:        0,
		HttpRequestSecret:  "",
		VerifyBlockParam:   nil,
		Metrics:            nil,
		Logger:             nil,
	}

	return New(cfg)
}

func createProdSrv() (*Server, error) {
	cfg := Config{
		StorageAdapter:     "",
		StorageURI:         "",
		StorageDB:          "",
		StorageIsFlush:     false,
		MinConn:            0,
		MaxConn:            0,
		KardiaURLs:         nil,
		KardiaTrustedNodes: nil,
		CacheAdapter:       "",
		CacheURL:           "",
		CacheDB:            0,
		CacheIsFlush:       false,
		BlockBuffer:        0,
		HttpRequestSecret:  "",
		VerifyBlockParam:   nil,
		Metrics:            nil,
		Logger:             nil,
	}

	return New(cfg)
}
