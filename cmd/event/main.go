package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/rpc"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/handler/event"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func main() {
	ctx := context.Background()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	serviceCfg, err := cfg.New()
	if err != nil {
		panic(err.Error())
	}

	logger, err := utils.NewLogger(serviceCfg)
	if err != nil {
		panic(err.Error())
	}

	node, err := kardia.NewNode(serviceCfg.KardiaWSNodes[0], logger)
	if err != nil {
		return
	}

	dbConfig := db.Config{
		DbAdapter: db.Adapter(serviceCfg.StorageDriver),
		DbName:    serviceCfg.StorageDB,
		URL:       serviceCfg.StorageURI,
		Logger:    logger,
		MinConn:   1,
		MaxConn:   4,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		panic(err)
	}

	cacheCfg := cache.Config{
		Adapter: cache.Adapter(serviceCfg.CacheEngine),
		URL:     serviceCfg.CacheURL,
		DB:      serviceCfg.CacheDB,
		Logger:  logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		panic(err)
	}

	eventCfg := event.Config{
		Node:   node,
		DB:     dbClient,
		Cache:  cacheClient,
		Logger: logger,
	}

	eventHandler, err := event.NewEventHandler(eventCfg)
	if err != nil {
		panic(err)
	}

	go func() {
		// todo: improve ping flow
		for {
			_, _ = node.LatestBlockNumber(ctx)
			time.Sleep(10 * time.Second)
		}
	}()

	poolSize := 4
	p, err := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		l := i.(*kardia.Log)
		if err := eventHandler.ProcessNewEventLog(ctx, l); err != nil {
			logger.Error("cannot process new event", zap.Error(err), zap.Any("Log", l))
			return
		}
	})
	// If process stop then release pool
	defer p.Release()

	for {
		logger.Debug("Start subscribe flow")
		time.Sleep(1 * time.Second)
		logEventCh := make(chan *kardia.Log, 10)
		sub, err := setupEventSubscription(ctx, node, logEventCh)
		if err != nil {
			// todo: handle close graceful
			logger.Debug("Cannot subscribe, closed", zap.Error(err))
			return
		}

		err = triggerEventWatcher(sub, logEventCh, p)
		if err != nil {
			logger.Debug("Somethings wrong, stop and try reconnect", zap.Error(err))
			sub.Unsubscribe()

		}
	}
}

func setupEventSubscription(ctx context.Context, n kardia.Node, channel interface{}) (*rpc.ClientSubscription, error) {
	args := kardia.FilterArgs{Address: []string{}}
	sub, err := n.KaiSubscribe(ctx, channel, "logs", args)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func triggerEventWatcher(sub *rpc.ClientSubscription, logsCh chan *kardia.Log, p *ants.PoolWithFunc) error {
	for {
		select {
		case err := <-sub.Err():
			// Handle error
			return err
		case log := <-logsCh:
			if err := p.Invoke(log); err != nil {
				fmt.Println("cannot invoke handler function", err)
			}

		}
	}
}
