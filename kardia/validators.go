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
	validators, err := node.Validators(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("TotalTime: ", time.Now().Sub(startLoadValidatorTime))
	fmt.Println("ValidatorInfo", validators)
	return nil, nil
}

func (ec *Client) getValidator(ctx context.Context, validatorSMCAddr string) (*types.Validator, error) {
	return nil, nil
}
