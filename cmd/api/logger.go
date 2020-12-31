// Package main
package main

import (
	"github.com/getsentry/sentry-go"
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
	sentryOpts := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, func(entry zapcore.Entry) error {
			e := sentry.NewEvent()
			e.Message = entry.Message
			switch entry.Level {
			case zap.InfoLevel:
				e.Level = sentry.LevelInfo
			case zap.DebugLevel:
				e.Level = sentry.LevelDebug
			case zap.WarnLevel:
				e.Level = sentry.LevelWarning
			case zap.ErrorLevel:
				e.Level = sentry.LevelError
			default:
				e.Level = sentry.LevelInfo
			}
			sentry.CaptureEvent(e)
			return nil
		})
	})

	return logCfg.Build(sentryOpts)
}
