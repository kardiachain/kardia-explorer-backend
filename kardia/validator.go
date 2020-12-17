/*
 *  Copyright 2020 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */

package kardia

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/go-kardia/lib/common"
	"go.uber.org/zap"
)

// GetValidatorInfo returns information of this validator
func (ec *Client) GetValidatorInfo(ctx context.Context, valSmcAddr common.Address) (*types.RPCValidator, error) {
	payload, err := ec.validatorUtil.Abi.Pack("inforValidator")
	if err != nil {
		ec.lgr.Error("Error packing validator info payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetValidatorInfo KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	var valInfo types.RPCValidator
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&valInfo, "inforValidator", res)
	if err != nil {
		ec.lgr.Error("Error unpacking validator info: ", zap.Error(err))
		return nil, err
	}
	rate, maxRate, maxChangeRate, err := ec.GetCommissionValidator(ctx, valSmcAddr)
	if err != nil {
		return nil, err
	}
	valInfo.CommissionRate = rate
	valInfo.MaxRate = maxRate
	valInfo.MaxChangeRate = maxChangeRate
	return &valInfo, nil
}

// GetDelegationRewards returns reward of a delegation
func (ec *Client) GetDelegationRewards(ctx context.Context, valSmcAddr common.Address, delegatorAddr common.Address) (*big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("getDelegationRewards", delegatorAddr)
	if err != nil {
		ec.lgr.Error("Error packing delegation rewards payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetDelegationRewards KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	var result struct {
		Rewards *big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "getDelegationRewards", res)
	if err != nil {
		ec.lgr.Error("Error unpacking delegation rewards: ", zap.Error(err))
		return nil, err
	}
	return result.Rewards, nil
}

// GetDelegatorStakedAmount returns staked amount of a delegator to current validator
func (ec *Client) GetDelegatorStakedAmount(ctx context.Context, valSmcAddr common.Address, delegatorAddr common.Address) (*big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("delegationByAddr", delegatorAddr)
	if err != nil {
		ec.lgr.Error("Error packing delegator staked amount payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetDelegatorStakedAmount KardiaCall error: ", zap.Error(err))
		return nil, err
	}

	var result struct {
		Stake          *big.Int
		PreviousPeriod *big.Int
		Height         *big.Int
		Shares         *big.Int
		Owner          common.Address
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "delegationByAddr", res)
	if err != nil {
		ec.lgr.Error("Error unpacking delegator's staked amount: ", zap.Error(err))
		return nil, err
	}
	return result.Stake, nil
}

// GetUDBEntry returns unbonded amount and withdrawable amount of a delegation
func (ec *Client) GetUDBEntries(ctx context.Context, valSmcAddr common.Address, delegatorAddr common.Address) (*big.Int, *big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("getUBDEntries", delegatorAddr)
	if err != nil {
		ec.lgr.Error("Error packing UDB entry payload: ", zap.Error(err))
		return nil, nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetUDBEntry KardiaCall error: ", zap.Error(err))
		return nil, nil, err
	}

	var result struct {
		Balances        []*big.Int
		CompletionTimes []*big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "getUBDEntries", res)
	if err != nil {
		ec.lgr.Error("Error unpacking UDB entry: ", zap.Error(err))
		return nil, nil, err
	}
	totalAmount := new(big.Int).SetInt64(0)
	withdrawableAmount := new(big.Int).SetInt64(0)
	now := new(big.Int).SetInt64(time.Now().Unix())
	for i, balance := range result.Balances {
		if result.CompletionTimes[i].Cmp(now) == -1 {
			withdrawableAmount = new(big.Int).Add(withdrawableAmount, balance)
		} else {
			totalAmount = new(big.Int).Add(totalAmount, balance)
		}
	}
	if totalAmount == nil || len(totalAmount.Bits()) == 0 {
		totalAmount = new(big.Int).SetInt64(0)
	}
	if withdrawableAmount == nil || len(withdrawableAmount.Bits()) == 0 {
		withdrawableAmount = new(big.Int).SetInt64(0)
	}
	return totalAmount, withdrawableAmount, nil
}

// GetMissedBlock returns missed block of this validator
func (ec *Client) GetMissedBlock(ctx context.Context, valAddr common.Address) ([]bool, error) {
	valSmcAddr, err := ec.GetValidatorContractFromOwner(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	payload, err := ec.validatorUtil.Abi.Pack("getMissedBlock")
	if err != nil {
		ec.lgr.Error("Error packing get missed blocks payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetUDBEntry KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	var result struct {
		MissedBlock []bool
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "getMissedBlock", res)
	if err != nil {
		ec.lgr.Error("Error unpack get missed blocks: ", zap.Error(err))
		return nil, err
	}
	return result.MissedBlock, nil
}

// GetValidator show info of a validator based on address
func (ec *Client) GetCommissionValidator(ctx context.Context, valSmcAddr common.Address) (*big.Int, *big.Int, *big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("commission")
	if err != nil {
		return nil, nil, nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		return nil, nil, nil, err
	}

	var result struct {
		Rate          *big.Int
		MaxRate       *big.Int
		MaxChangeRate *big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "commission", res)
	if err != nil {
		ec.lgr.Error("Error unpacking validator commission info", zap.Error(err))
		return nil, nil, nil, err
	}
	return result.Rate, result.MaxRate, result.MaxChangeRate, nil
}

// GetDelegators returns all delegators of a validator
func (ec *Client) GetDelegators(ctx context.Context, valSmcAddr common.Address) ([]*types.RPCDelegator, error) {
	payload, err := ec.validatorUtil.Abi.Pack("getDelegations")
	if err != nil {
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		return nil, err
	}

	var result struct {
		DelAddrs []common.Address
		Shares   []*big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "getDelegations", res)
	if err != nil {
		ec.lgr.Error("Error unpacking delegation details", zap.Error(err))
		return nil, err
	}
	var delegators []*types.RPCDelegator
	for _, delAddr := range result.DelAddrs {
		reward, err := ec.GetDelegationRewards(ctx, valSmcAddr, delAddr)
		if err != nil {
			return nil, err
		}
		stakedAmount, err := ec.GetDelegatorStakedAmount(ctx, valSmcAddr, delAddr)
		if err != nil {
			return nil, err
		}
		delegators = append(delegators, &types.RPCDelegator{
			Address:      delAddr,
			StakedAmount: stakedAmount,
			Reward:       reward,
		})
	}
	return delegators, nil
}

// GetSlashEvents returns detailed all slash events of this validator
func (ec *Client) GetSlashEvents(ctx context.Context, valAddr common.Address) ([]*types.SlashEvents, error) {
	var (
		one         = big.NewInt(1)
		slashEvents []*types.SlashEvents
	)
	valSmcAddr, err := ec.GetValidatorContractFromOwner(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	for i := new(big.Int).SetInt64(0); ; i.Add(i, one) {
		payload, err := ec.validatorUtil.Abi.Pack("slashEvents", i)
		if err != nil {
			return nil, err
		}
		res, err := ec.KardiaCall(ctx, ec.contructCallArgs(valSmcAddr.Hex(), payload))
		if err != nil {
			return nil, err
		}
		if strings.TrimRight(res.String(), "0") == "0x" {
			break
		}
		var result struct {
			Period   *big.Int
			Fraction *big.Int
			Height   *big.Int
		}
		// unpack result
		err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "slashEvents", res)
		if err != nil {
			ec.lgr.Error("Error unpacking slash event", zap.Error(err))
			return nil, err
		}
		slashEvents = append(slashEvents, &types.SlashEvents{
			Period:   result.Period.String(),
			Fraction: result.Fraction.String(),
			Height:   result.Height.String(),
		})
	}
	return slashEvents, nil
}
