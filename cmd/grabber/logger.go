// Package main
package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kardiachain/explorer-backend/cfg"
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
