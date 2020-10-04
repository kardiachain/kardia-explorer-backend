// Package main
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	DefaultFetchRate = 1 * time.Second
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger, err := zap.NewProduction()
	if err != nil {
		panic("cannot init logger")
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("cannot sync logger", zap.Error(err))
		}
	}()

	defer func() {
		if err := recover(); err != nil {
			logger.Error("cannot recover")
		}
	}()

	var rpcUrl string
	var checkTxCount bool
	var mongoUrl string
	var dbName string
	var startFrom uint64
	var blockRangeLimit uint64
	var workersCount uint
	var flushDb bool

	app := cli.NewApp()
	app.Usage = "Grabber populates a mongo database with explorer data."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "rpc-url, u",
			Value:       "http://dev-7.kardiachain.io",
			Usage:       "rpc api url",
			Destination: &rpcUrl,
		},
		cli.StringFlag{
			Name:        "mongo-url, m",
			Value:       "127.0.0.1:27017",
			Usage:       "mongo connection url",
			Destination: &mongoUrl,
		},
		cli.BoolFlag{
			Name:        "tx-count, tx",
			Usage:       "check a transactions count for every block(a heavy operation)",
			Destination: &checkTxCount,
		},
		cli.StringFlag{
			Name:        "mongo-dbname, db",
			Value:       "blockDB",
			Usage:       "mongo database name",
			Destination: &dbName,
		},
		cli.Uint64Flag{
			Name:        "start-from, s",
			Value:       0, //95730,
			Usage:       "refill from this block",
			Destination: &startFrom,
		},
		cli.Uint64Flag{
			Name:        "block-range-limit, b",
			Value:       10000,
			Usage:       "block range limit",
			Destination: &blockRangeLimit,
		},
		cli.UintFlag{
			Name:        "workers-amount, w",
			Value:       10,
			Usage:       "parallel workers amount",
			Destination: &workersCount,
		},
		cli.StringSliceFlag{
			Name:  "locked-accounts",
			Usage: "accounts with locked funds to exclude from rich list and circulating supply",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "Minimum log level to include. Lower levels will be discarded. (debug, info, warn, error, panic, fatal)",
		},
		cli.BoolFlag{
			Name:        "flush-db, f",
			Usage:       "Flush database",
			Destination: &flushDb,
		},
	}

	_, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
			cancel()
		}
	}()

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
	logger.Info("Stopping")
}

func setupAction(cfg zap.Config, c *cli.Context) error {
	//if c.IsSet("log-level") {
	//	var lvl zapcore.Level
	//	s := c.String("log-level")
	//	if err := lvl.Set(s); err != nil {
	//		return fmt.Errorf("invalid log-level %q: %v", s, err)
	//	}
	//	cfg.Level.SetLevel(lvl)
	//}
	//
	//importer, err := backend.NewBackend(ctx, mongoUrl, rpcUrl, dbName, lockedAccounts, nil, logger, nil, flushDb, metricsProvider)
	//if err != nil {
	//	return fmt.Errorf("failed to create backend: %v", err)
	//}
	//
	//listenerImporter, err := backend.NewBackend(ctx, mongoUrl, rpcUrl, dbName, lockedAccounts, nil, logger.With(zap.String("service", "listener")), nil, flushDb, metricsProvider)
	//if err != nil {
	//	return fmt.Errorf("failed to create backend: %v", err)
	//}
	//backfillImporter, err := backend.NewBackend(ctx, mongoUrl, rpcUrl, dbName, lockedAccounts, nil, logger.With(zap.String("service", "backfill")), nil, false, metricsProvider)
	//if err != nil {
	//	return fmt.Errorf("failed to create backend: %v", err)
	//}
	//
	//go listener(ctx, listenerImporter)
	//go backfill(ctx, backfillImporter, startFrom, checkTxCount)
	//
	////go migrator(ctx, importer, logger)
	//go updateStats(ctx, importer)
	//go updateAddresses(ctx, 3*time.Minute, false, blockRangeLimit, workersCount, importer) // update only addresses
	//updateAddresses(ctx, 5*time.Second, true, blockRangeLimit, workersCount, importer)     // update contracts
	return nil
}
