// Package kardia
package kardia

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/explorer-backend/types"
)

func convertValidatorInfo(val *types.Validator, totalStakedAmount *big.Int, status int) (*types.Validator, error) {
	var (
		err  error
		zero = new(big.Int).SetInt64(0)
	)
	if val.CommissionRate, err = convertBigIntToPercentage(val.CommissionRate); err != nil {
		return nil, err
	}
	if val.MaxRate, err = convertBigIntToPercentage(val.MaxRate); err != nil {
		return nil, err
	}
	if val.MaxChangeRate, err = convertBigIntToPercentage(val.MaxChangeRate); err != nil {
		return nil, err
	}
	if totalStakedAmount != nil && totalStakedAmount.Cmp(zero) == 1 && status == 2 {
		if val.VotingPowerPercentage, err = calculateVotingPower(val.StakedAmount, totalStakedAmount); err != nil {
			return nil, err
		}
	} else {
		val.VotingPowerPercentage = "0"
	}
	val.SigningInfo.IndicatorRate = 100 - float64(val.SigningInfo.MissedBlockCounter)/100
	return val, nil
}

func convertBigIntToPercentage(raw string) (string, error) {
	input, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", ErrParsingBigIntFromString
	}
	tmp := new(big.Int).Mul(input, tenPoweredBy18)
	result := new(big.Int).Div(tmp, tenPoweredBy18).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-16]+"."+result[len(result)-16:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func calculateVotingPower(raw string, total *big.Int) (string, error) {
	valStakedAmount, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", ErrParsingBigIntFromString
	}
	tmp := new(big.Int).Mul(valStakedAmount, tenPoweredBy5)
	result := new(big.Int).Div(tmp, total).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-3]+"."+result[len(result)-3:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func (ec *Client) getValidatorRole(valsSet []common.Address, address common.Address, status uint8) int {
	// if he's in validators set, he is a proposer
	for _, val := range valsSet {
		if val.Equal(address) {
			return 2
		}
	}
	// else if his node is started, he is a normal validator
	if status == 2 {
		return 1
	}
	// otherwise he is a candidate
	return 0
}
