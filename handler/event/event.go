// Package handler
package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

type Config struct {
	Node   kardia.Node
	DB     db.Client
	Cache  cache.Client
	Logger *zap.Logger
}

func NewEventHandler(cfg Config) (*Event, error) {
	e := &Event{
		node:   cfg.Node,
		db:     cfg.DB,
		cache:  cfg.Cache,
		logger: cfg.Logger.With(zap.String("Handler", "Event")),
	}
	return e, nil
}

type Event struct {
	node   kardia.Node
	db     db.Client
	cache  cache.Client
	logger *zap.Logger
}

func (h *Event) ProcessNewEventLog(ctx context.Context, l *kardia.Log) error {
	lgr := h.logger.With(zap.String("method", "ProcessNewEventLog"))
	lgr.Info("Process log event", zap.Any("Log", l))
	l.Address = common.HexToAddress(l.Address).String()
	tx, err := h.node.GetTransaction(ctx, l.TxHash)
	if err != nil {
		lgr.Error("Cannot get transaction", zap.Error(err))
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
	r, err := h.node.GetTransactionReceipt(ctx, l.TxHash)
	if err != nil {
		lgr.Error("cannot get receipt", zap.Error(err))
	}

	if r.ContractAddress != "0x" {
		// Insert new contract
		if err := h.db.InsertContract(ctx, &types.Contract{}, &types.Address{}); err != nil {
			lgr.Error("cannot insert new contract", zap.Error(err))
			return err
		}
	}

	getABITime := time.Now()
	// 1. Get ABI by address, if exist then process
	smcABI, err := h.getSMCAbi(ctx, l)
	if err == nil {
		decodedLog, err := kardia.UnpackLog(l, smcABI)
		if err != nil {
			lgr.Error("cannot decode log", zap.Error(err))
		}
		l = decodedLog
	}

	if errors.Is(err, types.ErrABINotFound) {
		// Check KRC

		// Not kind of KRC20

		if err := h.cache.UpdateSMCType(ctx, l.Address, types.TypeContractNormal); err != nil {
			lgr.Error("cannot update SMCType to Normal", zap.Error(err))
		}
	}

	if errors.Is(err, types.ErrSMCTypeNormal) {
		lgr.Debug("SMCType: Normal. Do nothing")
	}

	fmt.Println("SMC ABI", smcABI)
	lgr.Debug("Fetch ABI time", zap.Duration("Consumed", time.Since(getABITime)))

	if l.Topics[0] == cfg.KRCTransferTopic {
		if err := h.processKrcSMC(ctx, eventLog); err != nil {
			return err
		}
	}

	return h.db.InsertEvents([]types.Log{eventLog})
}

func (h *Event) processNormalSMC() error {
	return nil
}

func (h *Event) processKrcSMC(ctx context.Context, l types.Log) error {
	iTx := h.getInternalTxs(ctx, l)
	if iTx != nil {
		internalTxsList = append(internalTxsList, iTx)
	}
	holders, err := s.getKRCHolder(ctx, decodedLog)
	if err != nil {
		s.logger.Warn("Cannot get KRC holder", zap.Error(err), zap.Any("log", logs[i]))
		continue
	}
	holdersList = append(holdersList, holders...)
}

func (h *Event) getABIByAddress(ctx context.Context, l *kardia.Log) (*abi.ABI, error) {
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

	return nil, fmt.Errorf("%w", types.ErrABINotFound)
}

func (h *Event) getSMCAbi(ctx context.Context, log *kardia.Log) (*abi.ABI, error) {
	lgr := h.logger.With(zap.String("method", "getSMCAbi"))
	// Get ABI string from cache
	smcABIStr, err := h.cache.SMCAbi(ctx, log.Address)
	if err == nil {
		return utils.DecodeSMCABIFromBase64(smcABIStr)
	}

	smcType, err := h.cache.SMCType(ctx, log.Address)
	if err == nil {
		// Try to get ABI with SMCType
		smcABIStr, err = h.cache.SMCAbi(ctx, cfg.SMCTypePrefix+smcType)
		if err == nil {
			return utils.DecodeSMCABIFromBase64(smcABIStr)
		}

		// todo: Test unit this case
		if smcType == types.TypeContractNormal {
			return nil, fmt.Errorf("%w", types.ErrSMCTypeNormal)
		}
	}

	// Try to get SMC from database
	smc, _, err := h.db.Contract(ctx, log.Address)
	if err != nil {
		lgr.Error("Cannot get smc info from db", zap.Error(err), zap.String("smcAddr", log.Address))
		return nil, fmt.Errorf("%w", types.ErrABINotFound)
	}

	if smc.ABI != "" {
		return utils.DecodeSMCABIFromBase64(smc.ABI)
	}

	if smc.Type == "" {
		return nil, fmt.Errorf("%w", types.ErrABINotFound)
	}

	err = h.cache.UpdateSMCAbi(ctx, log.Address, cfg.SMCTypePrefix+smc.Type)
	if err != nil {
		lgr.Warn("Cannot store smc abi to cache", zap.Error(err))
		return nil, err
	}
	smcABIStr, err = h.cache.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
	if err == nil {
		return utils.DecodeSMCABIFromBase64(smcABIStr)
	}

	// query then reinsert abi of this SMC type to cache
	smcABIBase64, err := h.db.SMCABIByType(ctx, smc.Type)
	if err != nil {
		lgr.Warn("Cannot get smc abi by type from DB", zap.Error(err))
		return nil, err
	}
	err = h.cache.UpdateSMCAbi(ctx, cfg.SMCTypePrefix+smc.Type, smcABIBase64)
	if err != nil {
		lgr.Warn("Cannot store smc abi by type to cache", zap.Error(err))
		return nil, err
	}
	smcABIStr, err = h.cache.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
	if err != nil {
		lgr.Warn("Cannot get smc abi from cache", zap.Error(err))
		return nil, err
	}

	err = h.cache.UpdateSMCAbi(ctx, log.Address, smcABIStr)
	if err != nil {
		return nil, err
	}

	return utils.DecodeSMCABIFromBase64(smcABIStr)
}

func (h *Event) isKRC() {

}

func (h *Event) SyncKRC(ctx context.Context, l *types.Log) error {
	if l.Topics[0] != cfg.KRCTransferTopic {
		return nil
	}

	//var tokenTransfers []*types.TokenTransfer
	var tokenHolders []*types.TokenHolder

	for _, h := range tokenHolders {
		fmt.Println("h", h)
		return nil
	}

	return nil

}

func (h *Event) processHolders(ctx context.Context, holders []*types.TokenHolder) {
	for _, h := range holders {
		// Check if new address if contract
		fmt.Println("h", h)

	}
}

func (h *Event) getInternalTxs(ctx context.Context, log types.Log) *types.TokenTransfer {
	var (
		from, to, value string
		ok              bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return nil
	}
	value, ok = log.Arguments["value"].(string)
	if !ok {
		return nil
	}
	return &types.TokenTransfer{
		TransactionHash: log.TxHash,
		BlockHeight:     log.BlockHeight,
		Contract:        log.Address,
		From:            from,
		To:              to,
		Value:           value,
		Time:            log.Time,
		LogIndex:        log.Index,
	}
}
