// Package kardia
package kardia

import (
	"github.com/kardiachain/go-kaiclient/kardia"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func convertValidator(nValidator *kardia.Validator) (*types.Validator, error) {
	v := &types.Validator{
		Address:    nValidator.Signer.String(),
		SmcAddress: nValidator.SMCAddress.String(),
		Status:     nValidator.Status,
		Jailed:     nValidator.Jailed,
		Name:       validatorNameInString(nValidator.Name),
		//VotingPowerPercentage: nValidator.,
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
	// Calculate rate
	var err error
	if v.CommissionRate, err = convertBigIntToPercentage(nValidator.Commission.Rate.String()); err != nil {
		return nil, err
	}
	if v.MaxRate, err = convertBigIntToPercentage(nValidator.Commission.MaxRate.String()); err != nil {
		return nil, err
	}
	if v.MaxChangeRate, err = convertBigIntToPercentage(nValidator.Commission.MaxChangeRate.String()); err != nil {
		return nil, err
	}
	//if totalStakedAmount != nil && totalStakedAmount.Cmp(zero) == 1 && status == 2 {
	//	if val.VotingPowerPercentage, err = calculateVotingPower(val.StakedAmount, totalStakedAmount); err != nil {
	//		return nil, err
	//	}
	//} else {
	//	val.VotingPowerPercentage = "0"
	//}
	//val.SigningInfo.IndicatorRate = 100 - float64(val.SigningInfo.MissedBlockCounter)/100

	return v, nil
}
