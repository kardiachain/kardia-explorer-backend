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

// Package cfg
package cfg

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	ModeDev        = "dev"
	ModeProduction = "prod"
)

type ExplorerConfig struct {
	ServerMode        string
	Port              string
	HttpRequestSecret string

	LogLevel string

	IsReloadBootData bool

	DefaultAPITimeout     time.Duration
	DefaultBlockFetchTime time.Duration

	BufferedBlocks int64

	CoinMarketAPIKey string

	CacheEngine      string
	CacheURL         string
	CacheDB          int
	CacheFile        string
	CacheIsFlush     bool
	CachePassword    string
	CacheExpiredTime time.Duration

	KardiaProtocol     string
	KardiaPublicNodes  []string
	KardiaTrustedNodes []string
	KardiaWSNodes      []string

	StorageDriver  string
	StorageURI     string
	StorageDB      string
	StorageMinConn int
	StorageMaxConn int
	StorageIsFlush bool

	ListenerInterval time.Duration
	BackfillInterval time.Duration
	VerifierInterval time.Duration

	VerifyBlockParam *types.VerifyBlockParam

	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsSecretRegion    string

	UploaderBucket     string
	UploaderAcl        string
	UploaderKey        string
	UploaderPathAvatar string
}

func New() (ExplorerConfig, error) {
	isReloadBootDataStr := os.Getenv("IS_RELOAD_BOOT_DATA")
	isReloadBootData, err := strconv.ParseBool(isReloadBootDataStr)
	if err != nil {
		isReloadBootData = true
	}

	apiDefaultTimeoutStr := os.Getenv("DEFAULT_API_TIMEOUT")
	apiDefaultTimeout, err := strconv.Atoi(apiDefaultTimeoutStr)
	if err != nil {
		apiDefaultTimeout = 2
	}

	apiDefaultBlockFetchTimeStr := os.Getenv("DEFAULT_BLOCK_FETCH_TIME")
	apiDefaultBlockFetchTime, err := strconv.Atoi(apiDefaultBlockFetchTimeStr)
	if err != nil {
		apiDefaultBlockFetchTime = 500
	}

	cacheExpiredTimeStr := os.Getenv("CACHE_EXPIRED_TIME")
	cacheExpiredTime, err := strconv.Atoi(cacheExpiredTimeStr)
	if err != nil {
		cacheExpiredTime = 12
	}

	bufferBlocksStr := os.Getenv("BUFFER_BLOCKS")
	bufferBlocks, err := strconv.Atoi(bufferBlocksStr)
	if err != nil {
		bufferBlocks = 50
	}

	cacheDBStr := os.Getenv("CACHE_DB")
	cacheDB, err := strconv.Atoi(cacheDBStr)
	if err != nil {
		return ExplorerConfig{}, err
	}

	cacheIsFlushStr := os.Getenv("CACHE_IS_FLUSH")
	cacheIsFlush, err := strconv.ParseBool(cacheIsFlushStr)
	if err != nil {
		cacheIsFlush = true
	}

	var (
		kardiaTrustedNodes []string
		kardiaPublicNodes  []string
		kardiaWSNodes      []string
	)
	kardiaTrustedNodesStr := os.Getenv("KARDIA_TRUSTED_NODES")
	if kardiaTrustedNodesStr != "" {
		kardiaTrustedNodes = strings.Split(kardiaTrustedNodesStr, ",")
	} else {
		panic("missing trusted node URLs in config")
	}
	kardiaPublicNodesStr := os.Getenv("KARDIA_PUBLIC_NODES")
	if kardiaPublicNodesStr != "" {
		kardiaPublicNodes = strings.Split(kardiaPublicNodesStr, ",")
	} else {
		panic("missing public node URLs in config")
	}

	kardiaWSNodesStr := os.Getenv("KARDIA_WS_NODES")
	if kardiaWSNodesStr != "" {
		kardiaWSNodes = strings.Split(kardiaWSNodesStr, ",")
	} else {
		panic("missing websocket node URLs in config")
	}

	listenerIntervalStr := os.Getenv("LISTENER_INTERVAL")
	listenerInterval, err := time.ParseDuration(listenerIntervalStr)
	if err != nil {
		listenerInterval = 1 * time.Second
	}
	backfillIntervalStr := os.Getenv("BACKFILL_INTERVAL")
	backfillInterval, err := time.ParseDuration(backfillIntervalStr)
	if err != nil {
		backfillInterval = 2 * time.Second
	}
	verifierIntervalStr := os.Getenv("VERIFIER_INTERVAL")
	verifierInterval, err := time.ParseDuration(verifierIntervalStr)
	if err != nil {
		verifierInterval = 2 * time.Second
	}

	storageMinConnStr := os.Getenv("STORAGE_MIN_CONN")
	storageMinConn, err := strconv.Atoi(storageMinConnStr)
	if err != nil {
		storageMinConn = 8
	}

	storageMaxConnStr := os.Getenv("STORAGE_MAX_CONN")
	storageMaxConn, err := strconv.Atoi(storageMaxConnStr)
	if err != nil {
		storageMaxConn = 32
	}

	storageIsFlushStr := os.Getenv("STORAGE_IS_FLUSH")
	storageIsFLush, err := strconv.ParseBool(storageIsFlushStr)
	if err != nil {
		storageIsFLush = false
	}

	verifyTxCountStr := os.Getenv("VERIFY_TX_COUNT")
	verifyTxCount, err := strconv.ParseBool(verifyTxCountStr)
	if err != nil {
		verifyTxCount = true
	}
	verifyBlockHashStr := os.Getenv("VERIFY_BLOCK_HASH")
	verifyBlockHash, err := strconv.ParseBool(verifyBlockHashStr)
	if err != nil {
		verifyBlockHash = true
	}

	AwsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	AwsSecretRegion := os.Getenv("AWS_SECRET_REGION")

	UploaderBucket := os.Getenv("AWS_UPLOADER_BUCKET")
	UploaderAcl := os.Getenv("AWS_UPLOADER_ACL")
	UploaderKey := os.Getenv("AWS_UPLOADER_KEY")
	UploaderPathAvatar := os.Getenv("AWS_UPLOADER_PATH_AVATAR")

	cfg := ExplorerConfig{
		ServerMode:            os.Getenv("SERVER_MODE"),
		Port:                  os.Getenv("PORT"),
		HttpRequestSecret:     os.Getenv("HTTP_REQUEST_SECRET"),
		LogLevel:              os.Getenv("LOG_LEVEL"),
		IsReloadBootData:      isReloadBootData,
		DefaultAPITimeout:     time.Duration(apiDefaultTimeout) * time.Second,
		DefaultBlockFetchTime: time.Duration(apiDefaultBlockFetchTime) * time.Millisecond,
		BufferedBlocks:        int64(bufferBlocks),
		CoinMarketAPIKey:      os.Getenv("COIN_MARKET_API_KEY"),
		CacheEngine:           os.Getenv("CACHE_ENGINE"),
		CacheURL:              os.Getenv("CACHE_URI"),
		CachePassword:         os.Getenv("CACHE_PASSWORD"),
		CacheFile:             os.Getenv("CACHE_FILE"),
		CacheDB:               cacheDB,
		CacheExpiredTime:      time.Duration(cacheExpiredTime) * time.Hour,

		CacheIsFlush: cacheIsFlush,

		KardiaPublicNodes:  kardiaPublicNodes,
		KardiaTrustedNodes: kardiaTrustedNodes,
		KardiaWSNodes:      kardiaWSNodes,

		StorageDriver:  os.Getenv("STORAGE_DRIVER"),
		StorageURI:     os.Getenv("STORAGE_URI"),
		StorageDB:      os.Getenv("STORAGE_DB"),
		StorageMinConn: storageMinConn,
		StorageMaxConn: storageMaxConn,
		StorageIsFlush: storageIsFLush,

		ListenerInterval: listenerInterval,
		BackfillInterval: backfillInterval,
		VerifierInterval: verifierInterval,

		VerifyBlockParam: &types.VerifyBlockParam{
			VerifyTxCount:   verifyTxCount,
			VerifyBlockHash: verifyBlockHash,
		},
		AwsAccessKeyId:     AwsAccessKeyId,
		AwsSecretAccessKey: AwsSecretAccessKey,
		AwsSecretRegion:    AwsSecretRegion,

		UploaderBucket:     UploaderBucket,
		UploaderAcl:        UploaderAcl,
		UploaderKey:        UploaderKey,
		UploaderPathAvatar: UploaderPathAvatar,
	}

	return cfg, nil
}
