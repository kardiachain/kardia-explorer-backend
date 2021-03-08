// Package handler
package handler

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

type IStakingHandler interface {
	Validators(ctx context.Context) ([]*types.Validator, error)
	ReloadValidator(ctx context.Context, validatorSMC string) error
}

//Validator return list validator
func (h *handler) Validators(ctx context.Context) ([]*types.Validator, error) {
	return h.db.Validators(ctx, db.ValidatorsFilter{})
}

func (h *handler) ReloadValidator(ctx context.Context, validatorSMC string) error {
	// Get validator info
	lgr := h.logger.With(zap.String("method", "ReloadValidator"))
	nValidator, err := h.trustedNode.Validator(ctx, validatorSMC)
	if err != nil {
		lgr.Warn("cannot get validator info", zap.Error(err))
		return err
	}

	validator := &types.Validator{
		Address:               nValidator.Signer,
		SmcAddress:            nValidator.SMCAddress,
		Status:                nValidator.Status,
		Role:                  0,
		Jailed:                false,
		Name:                  "",
		VotingPowerPercentage: "",
		StakedAmount:          "",
		AccumulatedCommission: "",
		UpdateTime:            0,
		CommissionRate:        "",
		TotalDelegators:       0,
		MaxRate:               "",
		MaxChangeRate:         "",
		SigningInfo:           nil,
		Delegators:            nil,
	}

	if err := h.db.UpsertValidator(ctx, validator); err != nil {
		lgr.Warn("cannot upsert validator", zap.Error(err))
		return err
	}

	return nil
}
