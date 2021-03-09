// Package kardia
package kardia

import (
	"context"
	"fmt"
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
	for _, v := range nValidators {
		validators = append(validators, &types.Validator{
			Address:               v.Signer,
			SmcAddress:            v.SMCAddress,
			Status:                v.Status,
			Role:                  0,
			Jailed:                v.Jailed,
			Name:                  validatorNameInString(v.Name),
			VotingPowerPercentage: "",
			StakedAmount:          "",
			AccumulatedCommission: "",
			UpdateTime:            0,
			CommissionRate:        "",
			TotalDelegators:       0,
			MaxRate:               "",
			MaxChangeRate:         "",
			SigningInfo:           nil,
		})
	}
	return validators, nil
}

func (ec *Client) getValidator(ctx context.Context, validatorSMCAddr string) (*types.Validator, error) {
	return nil, nil
}
