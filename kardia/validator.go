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
	"sort"

	"github.com/kardiachain/go-kardia/lib/common"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (ec *Client) Validator(ctx context.Context, address string) (*types.Validator, error) {
	var validator *types.Validator
	err := ec.defaultClient.c.CallContext(ctx, &validator, "kai_validator", address, true)
	if err != nil {
		return nil, err
	}
	valsSet, err := ec.GetValidatorSets(ctx)
	if err != nil {
		return nil, err
	}
	// update validator's role
	validator.Role = ec.getValidatorRole(valsSet, common.HexToAddress(validator.Address), validator.Status)
	// calculate his rate from big.Int
	convertedVal, err := convertValidatorInfo(validator, nil, validator.Role)
	if err != nil {
		return nil, err
	}
	return convertedVal, nil
}

func (ec *Client) Validators(ctx context.Context) ([]*types.Validator, error) {
	var (
		proposersStakedAmount = big.NewInt(0)
		validators            []*types.Validator
	)
	err := ec.defaultClient.c.CallContext(ctx, &validators, "kai_validators", true)
	if err != nil {
		return nil, err
	}
	// compare staked amount btw validators to determine their status
	sort.Slice(validators, func(i, j int) bool {
		iAmount, _ := new(big.Int).SetString(validators[i].StakedAmount, 10)
		jAmount, _ := new(big.Int).SetString(validators[j].StakedAmount, 10)
		return iAmount.Cmp(jAmount) == 1
	})
	var (
		delegators                 = make(map[string]bool)
		totalProposers             = 0
		totalValidators            = 0
		totalCandidates            = 0
		totalStakedAmount          = big.NewInt(0)
		totalDelegatorStakedAmount = big.NewInt(0)

		valStakedAmount *big.Int
		delStakedAmount *big.Int
		ok              bool
	)
	valsSet, err := ec.GetValidatorSets(ctx)
	if err != nil {
		return nil, err
	}
	for _, val := range validators {
		for _, del := range val.Delegators {
			delegators[del.Address] = true
			// exclude validator self delegation
			if del.Address == val.Address {
				continue
			}
			delStakedAmount, ok = new(big.Int).SetString(del.StakedAmount, 10)
			if !ok {
				return nil, ErrParsingBigIntFromString
			}
			totalDelegatorStakedAmount = new(big.Int).Add(totalDelegatorStakedAmount, delStakedAmount)
		}
		valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
		if !ok {
			return nil, ErrParsingBigIntFromString
		}
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, valStakedAmount)
		val.Role = ec.getValidatorRole(valsSet, common.HexToAddress(val.Address), val.Status)
		// validator who started a node and not in validators set is a normal validator
		if val.Role == 2 {
			totalProposers++
			totalValidators++
			valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
			if !ok {
				return nil, ErrParsingBigIntFromString
			}
			proposersStakedAmount = new(big.Int).Add(proposersStakedAmount, valStakedAmount)
		} else if val.Role == 1 {
			totalValidators++
		} else if val.Role == 0 {
			totalCandidates++
		}
	}
	var returnValsList []*types.Validator
	for _, val := range validators {
		convertedVal, err := convertValidatorInfo(val, proposersStakedAmount, val.Role)
		if err != nil {
			return nil, err
		}
		returnValsList = append(returnValsList, convertedVal)
	}

	return returnValsList, nil
}

// GetValidatorInfo returns information of this validator
func (ec *Client) GetValidatorInfo(ctx context.Context, valSmcAddr common.Address) (*types.RPCValidator, error) {
	payload, err := ec.validatorUtil.Abi.Pack("inforValidator")
	if err != nil {
		ec.lgr.Error("Error packing validator info payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
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
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
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
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
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
func (ec *Client) GetUDBEntries(ctx context.Context, valSmcAddr common.Address, delegatorAddr common.Address) ([]*UnbondedRecord, error) {
	payload, err := ec.validatorUtil.Abi.Pack("getUBDEntries", delegatorAddr)
	if err != nil {
		ec.lgr.Error("Error packing UDB entry payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetUDBEntry KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var records []*UnbondedRecord
	var result struct {
		Balances        []*big.Int
		CompletionTimes []*big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "getUBDEntries", res)
	if err != nil {
		ec.lgr.Error("Error unpacking UDB entry: ", zap.Error(err))
		return nil, err
	}

	for id := range result.CompletionTimes {
		records = append(records, &UnbondedRecord{
			Balance:        result.Balances[id],
			CompletionTime: result.CompletionTimes[id],
		})
	}
	return records, nil
}

// GetSigningInfo returns signing info of this validator
func (ec *Client) GetSigningInfo(ctx context.Context, valSmcAddr common.Address) (*types.SigningInfo, error) {
	payload, err := ec.validatorUtil.Abi.Pack("signingInfo")
	if err != nil {
		ec.lgr.Error("Error packing get signingInfo payload: ", zap.Error(err))
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetSigningInfo KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	var result struct {
		StartHeight        *big.Int
		IndexOffset        *big.Int
		Tombstoned         bool
		MissedBlockCounter *big.Int
		JailedUntil        *big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&result, "signingInfo", res)
	if err != nil {
		ec.lgr.Error("Error unpack get signingInfo: ", zap.Error(err))
		return nil, err
	}
	return &types.SigningInfo{
		StartHeight:        result.StartHeight.Uint64(),
		IndexOffset:        result.IndexOffset.Uint64(),
		Tombstoned:         result.Tombstoned,
		MissedBlockCounter: result.MissedBlockCounter.Uint64(),
		JailedUntil:        result.JailedUntil.Uint64(),
	}, nil
}

// GetValidator show info of a validator based on address
func (ec *Client) GetCommissionValidator(ctx context.Context, valSmcAddr common.Address) (*big.Int, *big.Int, *big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("commission")
	if err != nil {
		return nil, nil, nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
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
	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
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
			continue
		}
		stakedAmount, err := ec.GetDelegatorStakedAmount(ctx, valSmcAddr, delAddr)
		if err != nil {
			continue
		}
		delegators = append(delegators, &types.RPCDelegator{
			Address:      delAddr,
			StakedAmount: stakedAmount,
			Reward:       reward,
		})
	}
	return delegators, nil
}

// GetSlashEventsLength returns number of slash events of this validator
func (ec *Client) GetSlashEventsLength(ctx context.Context, valSmcAddr common.Address) (*big.Int, error) {
	payload, err := ec.validatorUtil.Abi.Pack("getSlashEventsLength")
	if err != nil {
		ec.lgr.Error("Error packing get slash events length payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("GetSlashEventsLength KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var slashEventsLength *big.Int
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&slashEventsLength, "getSlashEventsLength", res)
	if err != nil {
		ec.lgr.Error("Error unpacking get slash events length error: ", zap.Error(err))
		return nil, err
	}
	return slashEventsLength, nil
}

// GetSlashEvents returns detailed all slash events of this validator
func (ec *Client) GetSlashEvents(ctx context.Context, valAddr common.Address) ([]*types.SlashEvents, error) {
	var (
		one         = big.NewInt(1)
		slashEvents []*types.SlashEvents
	)
	valSmcAddr, err := ec.GetValidatorSMCFromOwner(ctx, valAddr)
	if err != nil || valSmcAddr.Equal(common.Address{}) {
		ec.lgr.Error("Error getting validator contract address: ", zap.Any("valSmcAddr", valSmcAddr), zap.Error(err))
		return nil, err
	}
	length, err := ec.GetSlashEventsLength(ctx, valSmcAddr)
	if length == nil {
		return nil, nil
	}
	if err != nil {
		ec.lgr.Error("Error getting slash events length: ", zap.Any("valSmcAddr", valSmcAddr), zap.Error(err))
		return nil, err
	}
	for i := new(big.Int).SetInt64(0); i.Cmp(length) < 0; i.Add(i, one) {
		payload, err := ec.validatorUtil.Abi.Pack("slashEvents", i)
		if err != nil {
			return nil, err
		}
		res, err := ec.KardiaCall(ctx, contructCallArgs(valSmcAddr.Hex(), payload))
		if err != nil {
			ec.lgr.Warn("GetSlashEvents KardiaCall Error: ", zap.String("i", i.String()), zap.String("payload", common.Bytes(payload).String()), zap.Error(err))
			return nil, err
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
