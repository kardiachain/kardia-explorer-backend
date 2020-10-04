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
// Package features
package features_test

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cfg"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	if err := godotenv.Load(); err != nil {
		return
	}
}

type Response struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type suite struct {
	logger *zap.Logger

	resp *http.Response
	data Response

	db *pgxpool.Pool

	cfg       cfg.ExplorerConfig
	host      string
	urlParams []interface{}

	StepState
}
type StepState struct {
}

func (c *suite) SetupSuite() {

}

func FeatureContext(s *godog.ScenarioContext) {
	explorerCfg := cfg.ExplorerConfig{}
	logger, err := zap.NewProduction()
	if err != nil {
		return
	}
	db := NewDBPool(explorerCfg)
	c := &suite{
		logger: logger.With(zap.String("cmd", "BDD")),
		db:     db,
	}

	s.AfterScenario(func(sc *godog.Scenario, err error) {})
	c.SetupSuite()
}

func NewDBPool(cfg cfg.ExplorerConfig) *pgxpool.Pool {

	config, err := pgxpool.ParseConfig(cfg.PostgresURI)
	if err != nil {
		panic(err.Error())
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err.Error())
	}

	return pool
}

const (
	ServerDev  = "dev"
	ServerProd = "prod"
)
