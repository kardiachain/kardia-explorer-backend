// Package main
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	serviceCfg, err := cfg.New()
	if err != nil {
		panic(err.Error())
	}

	zap.L().Info("Start subscribe event")
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	waitExit := make(chan bool)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
			cancel()
			waitExit <- true
		}
	}()
	//go subscribeNewBlock(ctx)
	go runStakingSubscriber(ctx, serviceCfg)

	<-waitExit
	zap.L().Info("Stopped")
}
