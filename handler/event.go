// Package handler
package handler

import (
	"context"
	"errors"
	"time"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

type IEvent interface {
	ProcessNewEventLog(ctx context.Context)
}

func (h *handler) ProcessNewEventLog(ctx context.Context) {
	lgr := h.logger.With(zap.String("method", "ProcessNewEventLog"))
	wsNode := h.w.WSNode()
	args := kardia.FilterArgs{}
	logEventCh := make(chan *kardia.Log)
	sub, err := wsNode.KaiSubscribe(context.Background(), logEventCh, "logs", args)
	if err != nil {
		return
	}

	for {
		select {
		case err := <-sub.Err():
			lgr.Debug("subscribe err", zap.Error(err))
		case log := <-logEventCh:
			lgr.Debug("Log", zap.Any("detail", log))
			if err := h.onNewLogEvent(ctx, log); err != nil {
				return
			}
		}
	}
}

func (h *handler) onNewLogEvent(ctx context.Context, l *kardia.Log) error {
	lgr := h.logger.With(zap.String("method", "onNewLogEvent"))
	tx, err := h.w.TrustedNode().GetTransaction(ctx, l.TxHash)
	if err != nil {
		return err
	}
	eventLog := types.Log{
		Address:     l.Address,
		Arguments:   nil,
		Topics:      l.Topics,
		Data:        l.Data,
		BlockHeight: l.BlockHeight,
		Time:        tx.Time,
		TxHash:      l.TxHash,
		TxIndex:     l.TxIndex,
		BlockHash:   l.BlockHash,
		Index:       l.Index,
		Removed:     l.Removed,
	}

	getABITime := time.Now()
	// 1. Get ABI by address, if exist then process

	smcABI, err := h.getSMCAbi(ctx, l)
	if err != nil {
		return err
	}
	lgr.Debug("Fetch ABI time", zap.Duration("Consumed", time.Since(getABITime)))

	if l.Topics[0] == cfg.KRCTransferTopic {

	}

	return h.db.InsertEvents([]types.Log{eventLog})
}

func (h *handler) processNormalSMC() error {

}

func (h *handler) processKrcSMC() error {

}

func (h *handler) getABIByAddress(ctx context.Context, l *kardia.Log) (*abi.ABI, error) {
	// Get ABI string from cache
	smcABIStr, err := h.cache.SMCAbi(ctx, l.Address)
	if err == nil {
		return utils.DecodeSMCABIFromBase64(smcABIStr)
	}

	smc, _, err := h.db.Contract(ctx, l.Address)
	if err != nil {
		h.logger.Warn("Cannot get smc info from db", zap.Error(err), zap.String("smcAddr", l.Address))
		return nil, err
	}

	if smc.ABI != "" {
		return utils.DecodeSMCABIFromBase64(smc.ABI)
	}

	if smc.Type == "" {
		return nil, errors.New("cannot find SMC ABI")
	}

}

func (h *handler) getSMCAbi(ctx context.Context, log *kardia.Log) (*abi.ABI, error) {
	// Get ABI string from cache
	smcABIStr, err := h.cache.SMCAbi(ctx, log.Address)
	if err == nil {
		return utils.DecodeSMCABIFromBase64(smcABIStr)
	}

	smc, _, err := h.db.Contract(ctx, log.Address)
	if err != nil {
		h.logger.Warn("Cannot get smc info from db", zap.Error(err), zap.String("smcAddr", log.Address))
		return nil, err
	}

	if smc.ABI != "" {
		return utils.DecodeSMCABIFromBase64(smc.ABI)
	}

	if smc.Type == "" {
		return nil, errors.New("cannot find SMC ABI")
	}

	err = h.cache.UpdateSMCAbi(ctx, log.Address, cfg.SMCTypePrefix+smc.Type)
	if err != nil {
		h.logger.Warn("Cannot store smc abi to cache", zap.Error(err))
		return nil, err
	}
	smcABIStr, err = h.cache.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
	if err == nil {
		return utils.DecodeSMCABIFromBase64(smcABIStr)
	}

	// query then reinsert abi of this SMC type to cache
	smcABIBase64, err := h.db.SMCABIByType(ctx, smc.Type)
	if err != nil {
		h.logger.Warn("Cannot get smc abi by type from DB", zap.Error(err))
		return nil, err
	}
	err = h.cache.UpdateSMCAbi(ctx, cfg.SMCTypePrefix+smc.Type, smcABIBase64)
	if err != nil {
		h.logger.Warn("Cannot store smc abi by type to cache", zap.Error(err))
		return nil, err
	}
	smcABIStr, err = h.cache.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
	if err != nil {
		h.logger.Warn("Cannot get smc abi from cache", zap.Error(err))
		return nil, err
	}

	err = h.cache.UpdateSMCAbi(ctx, log.Address, smcABIStr)
	if err != nil {
		return nil, err
	}

	//
	return utils.DecodeSMCABIFromBase64(smcABIStr)
}

func (h *handler) isKRC() {

}

func (h *handler) SyncKRC(ctx context.Context, l *types.Log) error {
	if l.Topics[0] != cfg.KRCTransferTopic {
		return nil
	}

	var tokenTransfers []*types.TokenTransfer
	var tokenHolders []*types.TokenHolder

	for _, h := range tokenHolders {

	}

}
