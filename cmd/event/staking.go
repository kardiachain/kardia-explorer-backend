package main

import (
	"context"

	"github.com/kardiachain/go-kaiclient/kardia"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/handler"
	"github.com/kardiachain/kardia-explorer-backend/server"
)

const (
	WSNode   = "ws://ws-dev.kardiachain.io/ws"
	IsReload = false
)

func subscribeStakingEvent(ctx context.Context, srv *server.Server) error {
	//Staking SMC subscribe
	lgr := srv.Logger.With(zap.String("method", "subscribeStakingEvent"))
	node, err := kardia.NewNode(WSNode, zap.L())
	if err != nil {
		return err
	}

	args := kardia.FilterArgs{Address: []string{cfg.StakingContractAddr}}
	//Validators SMC subscribe
	eventLogCh := make(chan *kardia.FilterLogs)
	sub, err := node.KaiSubscribe(ctx, eventLogCh, "logs", args)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-sub.Err():
			lgr.Error("Subscribe error", zap.Error(err))
		case l := <-eventLogCh:
			lgr.Info("Event", zap.Any("log", l))
		}
	}
}

func subscribeValidatorEvent(ctx context.Context, h handler.Handler) error {
	lgr := zap.L().With(zap.String("method", "subscribeValidatorEvent"))
	// Get list validators, include candidate
	validators, err := h.Validators(ctx)
	if err != nil {
		lgr.Warn("cannot get validators", zap.Error(err))
	}
	if err != nil {
		return err
	}
	node, err := kardia.NewNode(WSNode, lgr)
	if err != nil {
		return err
	}

	// Build list SMCAddr
	var validatorsSMCAddresses []string
	for _, v := range validators {
		validatorsSMCAddresses = append(validatorsSMCAddresses, v.SmcAddress.Hex())
	}
	args := kardia.FilterArgs{Address: validatorsSMCAddresses}
	//Validators SMC subscribe
	eventLogCh := make(chan *kardia.FilterLogs)
	sub, err := node.KaiSubscribe(ctx, eventLogCh, "logs", args)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-sub.Err():
			lgr.Error("Subscribe error", zap.Error(err))
		case l := <-eventLogCh:
			lgr.Info("Event", zap.Any("log", l))
		}
	}
}
