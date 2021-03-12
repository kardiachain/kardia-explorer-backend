// Package handler
package handler

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kardiachain/go-kaiclient/kardia"
	ctypes "github.com/kardiachain/go-kardia/types"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

type IStakingHandler interface {
	SubscribeStakingEvent(ctx context.Context) error
	SubscribeValidatorEvent(ctx context.Context) error
}

func (h *handler) SubscribeStakingEvent(ctx context.Context) error {
	//Staking SMC subscribe
	lgr := h.logger.With(zap.String("method", "subscribeStakingEvent"))
	wsNode := h.w.WSNode()

	args := kardia.FilterArgs{Address: []string{cfg.StakingContractAddr}}
	//Validators SMC subscribe
	eventLogCh := make(chan *kardia.FilterLogs)
	sub, err := wsNode.KaiSubscribe(ctx, eventLogCh, "logs", args)
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

func (h *handler) SubscribeValidatorEvent(ctx context.Context) error {
	lgr := h.logger.With(zap.String("method", "SubscribeValidatorEvent"))
	wsNode := h.w.WSNode()
	trustedNode := h.w.TrustedNode()
	nValidatorSMCAddresses, err := trustedNode.ValidatorSMCAddresses(ctx)
	if err != nil {
		lgr.Warn("cannot get validatorSMCAddresses", zap.Error(err))
	}
	var validatorsSMCAddresses []string
	for _, addr := range nValidatorSMCAddresses {
		validatorsSMCAddresses = append(validatorsSMCAddresses, addr.Hex())
	}
	headersCh := make(chan *ctypes.Header)
	sub, err := wsNode.SubscribeNewHead(context.Background(), headersCh)

	for {
		select {
		case err := <-sub.Err():
			lgr.Error("Subscribe error", zap.Error(err))
		case header := <-headersCh:
			h.processHeader(ctx, header)
		}
	}
}

func (h *handler) reloadValidator(ctx context.Context, validatorSMCAddress string) {
	lgr := h.logger.With(zap.String("method", "reloadValidator"))
	v, err := h.w.Validator(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Warn("cannot get validator info", zap.String("SMCAddress", validatorSMCAddress), zap.Error(err))
		return
	}
	lgr.Debug("ValidatorInfo", zap.Any("validator", v))
	// 1. Upsert validator with new info
	if err := h.db.UpsertValidator(ctx, v); err != nil {
		lgr.Error("cannot upsert validator", zap.Error(err))
		return
	}

}

func (h *handler) processValidatorEvent(ctx context.Context, l *kardia.FilterLogs) {

}

func (h *handler) processHeader(ctx context.Context, header *ctypes.Header) {
	lgr := h.logger.With(zap.String("method", "processHeader"))
	if header.NumTxs == 0 {
		lgr.Debug("block has no txs", zap.String("hash", header.Hash().Hex()))
		return
	}
	block, err := h.w.TrustedNode().BlockByHash(ctx, header.Hash().Hex())
	if err != nil {
		lgr.Debug("cannot get block", zap.Error(err))
		return
	}

	validatorAddresses, err := h.w.TrustedNode().ValidatorSMCAddresses(ctx)
	if err != nil {
		lgr.Debug("cannot check get validator addresses", zap.Any("address", validatorAddresses))
	}
	validatorMap := make(map[string]bool)
	for _, addr := range validatorAddresses {
		validatorMap[addr.Hex()] = true
	}

	for _, tx := range block.Txs {
		isExist, ok := validatorMap[tx.To]
		if !ok || isExist == false {
			continue
		}
		h.reloadValidator(ctx, tx.To)
		h.reloadDelegator(ctx, tx.To, tx.From)

		// 2. Calculate new stats
		if err := h.calculateStakingStats(ctx); err != nil {
			return
		}
	}

}

func (h *handler) reloadDelegator(ctx context.Context, validatorSMCAddress, delegatorAddress string) {
	lgr := h.logger.With(zap.String("method", "reloadDelegator"))
	d, err := h.w.Delegator(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		lgr.Warn("cannot get validator info", zap.String("SMCAddress", validatorSMCAddress), zap.Error(err))
		return
	}

	// Update delegator info
	lgr.Debug("DelegatorInfo", zap.Any("delegator", d))
	if err := h.db.UpsertDelegator(ctx, d); err != nil {
		lgr.Error("cannot upsert delegator", zap.Error(err))
		return
	}

}

func (h *handler) calculateStakingStats(ctx context.Context) error {
	lgr := h.logger.With(zap.String("method", "calculateStakingStats"))
	// Reload from db
	validators, err := h.db.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		lgr.Error("cannot load validator from storage", zap.Error(err))
		return err
	}

	// Get new total staked
	node := h.w.TrustedNode()
	totalStaked, err := node.TotalStakedAmount(ctx)
	if err != nil {
		lgr.Error("cannot get total staked amount", zap.Error(err))
		return err
	}
	fmt.Println("TotalStakedAmount: ", totalStaked.String())
	for id, v := range validators {
		votingPower, err := utils.CalculateVotingPower(v.StakedAmount, totalStaked)
		if err != nil {
			return err
		}
		validators[id].VotingPowerPercentage = votingPower
	}
	if err := h.db.UpsertValidators(ctx, validators); err != nil {
		return err
	}

	var validatorAddresses []string
	for _, v := range validators {
		validatorAddresses = append(validatorAddresses, v.Address)
	}
	totalValidator, err := h.db.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return err
	}
	totalProposer, err := h.db.Validators(ctx, db.ValidatorsFilter{Role: cfg.RoleProposer})
	if err != nil {
		return err
	}
	totalCandidate, err := h.db.Validators(ctx, db.ValidatorsFilter{Role: cfg.RoleCandidate})
	if err != nil {
		return err
	}
	totalUniqueDelegator, err := h.db.UniqueDelegators(ctx)
	if err != nil {
		return err
	}

	totalValidatorStakedAmount, err := h.db.GetStakedOfAddresses(ctx, validatorAddresses)
	if err != nil {
		return err
	}
	stakedAmountBigInt, ok := new(big.Int).SetString(totalValidatorStakedAmount, 10)
	if !ok {
		return fmt.Errorf("cannot load validator staked amount")
	}
	TotalDelegatorsStakedAmount := new(big.Int).Sub(totalStaked, stakedAmountBigInt)

	stats := &types.StakingStats{
		TotalValidators:            len(totalValidator) + len(totalProposer),
		TotalProposers:             len(totalProposer),
		TotalCandidates:            len(totalCandidate),
		TotalDelegators:            totalUniqueDelegator,
		TotalStakedAmount:          totalStaked.String(),
		TotalValidatorStakedAmount: totalValidatorStakedAmount,
		TotalDelegatorStakedAmount: TotalDelegatorsStakedAmount.String(),
	}

	if err := h.cache.UpdateStakingStats(ctx, stats); err != nil {
		return err
	}

	return nil
}
