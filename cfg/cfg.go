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

	"github.com/joho/godotenv"
)

type ExplorerConfig struct {
	ServerMode string
	Port       string

	BufferedBlocks int

	CacheEngine string
	CacheUrl    string
	CacheFile   string

	KardiaProtocol string
	KardiaURLs     []string

	PostgresUri     string
	PostgresDB      string
	PostgresMaxConn int

	MongoURL string
	MongoDB  string
}

func New() (ExplorerConfig, error) {
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	bufferedBlocksStr := os.Getenv("BUFFER_BLOCKS")
	bufferedBlocks, err := strconv.Atoi(bufferedBlocksStr)
	if err != nil {
		return ExplorerConfig{}, err
	}

	postgresMaxConnStr := os.Getenv("POSTGRES_MAX_CONN")
	postgresMaxConn, err := strconv.Atoi(postgresMaxConnStr)
	if err != nil {
		return ExplorerConfig{}, err
	}

	cfg := ExplorerConfig{
		ServerMode:      os.Getenv("SERVER_MODE"),
		Port:            os.Getenv("PORT"),
		BufferedBlocks:  bufferedBlocks,
		CacheEngine:     os.Getenv("CACHE_ENGINE"),
		CacheUrl:        os.Getenv("CACHE_URL"),
		CacheFile:       os.Getenv("CACHE_FILE"),
		KardiaProtocol:  os.Getenv("KARDIA_PROTOCOL"),
		KardiaURLs:      strings.Split(os.Getenv("KARDIA_URL"), ","),
		PostgresUri:     os.Getenv("POSTGRES_URI"),
		PostgresDB:      os.Getenv("POSTGRES_URI"),
		PostgresMaxConn: postgresMaxConn,
		MongoURL:        os.Getenv("MONGO_URL"),
		MongoDB:         os.Getenv("MONGO_DB"),
	}

	return cfg, nil
}
