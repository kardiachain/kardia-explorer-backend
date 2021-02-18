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
// Package main
package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
)

func newLogger(sCfg cfg.ExplorerConfig) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	switch sCfg.ServerMode {
	case cfg.ModeDev:
		logCfg = zap.NewDevelopmentConfig()
		logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	case cfg.ModeProduction:
		logCfg = zap.NewProductionConfig()
	}

	switch sCfg.LogLevel {
	case "info":
		logCfg.Level.SetLevel(zapcore.InfoLevel)
	case "debug":
		logCfg.Level.SetLevel(zapcore.DebugLevel)
	case "warn":
		logCfg.Level.SetLevel(zapcore.WarnLevel)
	default:
		logCfg.Level.SetLevel(zapcore.InfoLevel)
	}

	return logCfg.Build()
}
