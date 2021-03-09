// Package kardia
package kardia

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/kardiachain/go-kaiclient/kardia"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	nodeURL = "https://rpc.kardiachain.io"
)

func (ec *Client) getValidators(ctx context.Context) ([]*types.Validator, error) {
	node, err := kardia.NewNode(nodeURL, ec.lgr)
	if err != nil {
		return nil, err
	}
	startLoadValidatorTime := time.Now()
	nValidators, err := node.Validators(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("TotalTime: ", time.Now().Sub(startLoadValidatorTime))
	var validators []*types.Validator
	totalStakedAmount := big.NewInt(0)
	for _, v := range nValidators {
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, v.Tokens)
	}
	for _, v := range nValidators {
		commissionRate, err := convertBigIntToPercentage(v.Commission.Rate.String())
		if err != nil {
			return nil, err
		}
		commissionMaxChangeRate, err := convertBigIntToPercentage(v.Commission.MaxChangeRate.String())
		if err != nil {
			return nil, err
		}
		commissionMaxRate, err := convertBigIntToPercentage(v.Commission.MaxRate.String())
		if err != nil {
			return nil, err
		}
		votingPowerPercent, err := calculateVotingPower(v.Tokens.String(), totalStakedAmount)
		if err != nil {
			return nil, err
		}

		validators = append(validators, &types.Validator{
			Address:               v.Signer,
			SmcAddress:            v.SMCAddress,
			Status:                v.Status,
			Role:                  0,
			Jailed:                v.Jailed,
			Name:                  validatorNameInString(v.Name),
			VotingPowerPercentage: votingPowerPercent,
			StakedAmount:          v.Tokens.String(),
			AccumulatedCommission: v.AccumulatedCommission.String(),
			UpdateTime:            v.UpdateTime.Uint64(),
			CommissionRate:        commissionRate,
			MaxRate:               commissionMaxRate,
			MaxChangeRate:         commissionMaxChangeRate,
			SigningInfo: &types.SigningInfo{
				StartHeight:        v.SigningInfo.StartHeight.Uint64(),
				IndexOffset:        v.SigningInfo.IndexOffset.Uint64(),
				Tombstoned:         v.SigningInfo.Tombstoned,
				MissedBlockCounter: v.SigningInfo.MissedBlockCounter.Uint64(),
				IndicatorRate:      100 - float64(v.SigningInfo.MissedBlockCounter.Uint64())/100,
				JailedUntil:        v.SigningInfo.JailedUntil.Uint64(),
			},
		})
	}
	return validators, nil
}

func (ec *Client) getValidator(ctx context.Context, validatorSMCAddr string) (*types.Validator, error) {
	return nil, nil
}
