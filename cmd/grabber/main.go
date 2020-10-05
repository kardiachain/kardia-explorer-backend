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

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
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

	logger.Info("Start grabber...")

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

	app.Action = func(c *cli.Context) error {
		ctx := context.Background()
		srvConfig := server.Config{
			DBAdapter:       db.MGO,
			DBUrl:           mongoUrl,
			KardiaProtocol:  kardia.RPCProtocol,
			KardiaURL:       rpcUrl,
			CacheAdapter:    cache.Redis,
			CacheURL:        "",
			LockedAccount:   nil,
			Signers:         nil,
			IsFlushDatabase: false,
			Metrics:         nil,
			Logger:          nil,
		}
		srv, err := server.New(srvConfig)
		if err != nil {
			logger.Panic(err.Error())
		}

		// Start listener in new go routine
		// todo @longnd: Running multi goroutine same time
		go listener(ctx, srv)
		updateAddresses(ctx, true, 0, srv)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
	logger.Info("Stopping")
}
