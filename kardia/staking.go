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
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/go-kardia/lib/common"
	staking "github.com/kardiachain/go-kardia/mainchain/staking"
)

func (ec *Client) GetValidatorsByDelegator(ctx context.Context, delAddr common.Address) ([]*types.ValidatorsByDelegator, error) {
	// construct input data
	payload, err := ec.stakingUtil.Abi.Pack("getValidatorsByDelegator", delAddr)
	if err != nil {
		return nil, err
	}
	// make static call through `kai_kardiaCall` API
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(ec.stakingUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return nil, err
	}
	// get validators list of delegator
	var valAddrs struct {
		ValAddrs []common.Address
	}
	// unpacking data
	err = ec.stakingUtil.Abi.UnpackIntoInterface(&valAddrs, "getValidatorsByDelegator", res)
	if err != nil {
		return nil, err
	}

	// gather additional information about validators
	var valsList []*types.ValidatorsByDelegator
	for _, val := range valAddrs.ValAddrs {
		valInfo, err := ec.GetValidatorInfo(ctx, val)
		if err != nil {
			return nil, err
		}
		var name []byte
		for _, b := range valInfo.Name {
			if b != 0 {
				name = append(name, byte(b))
			}
		}
		owner, err := ec.GetValidatorContractFromOwner(ctx, val)
		if err != nil {
			return nil, err
		}
		reward, err := ec.GetDelegationRewards(ctx, val, delAddr)
		if err != nil {
			return nil, err
		}
		stakedAmount, err := ec.GetDelegatorStakedAmount(ctx, val, delAddr)
		if err != nil {
			return nil, err
		}
		unbondedAmount, withdrawableAmount, err := ec.GetUDBEntries(ctx, val, delAddr)
		if err != nil {
			return nil, err
		}
		validator := &types.ValidatorsByDelegator{
			Name:                  string(name),
			Validator:             owner,
			ValidatorContractAddr: val,
			ValidatorStatus:       valInfo.Status,
			StakedAmount:          stakedAmount.String(),
			ClaimableRewards:      reward.String(),
			UnbondedAmount:        unbondedAmount.String(),
			WithdrawableAmount:    withdrawableAmount.String(),
		}
		valsList = append(valsList, validator)
	}
	return valsList, nil
}

// GetValidatorInfo returns information of this validator
func (ec *Client) GetValidatorInfo(ctx context.Context, valSmcAddr common.Address) (*staking.Validator, error) {
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
	var valInfo staking.Validator
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&valInfo, "inforValidator", res)
	if err != nil {
		ec.lgr.Error("Error unpacking validator info: ", zap.Error(err))
		return nil, err
	}
	return &valInfo, nil
}

// GetValidatorContractFromOwner returns validator contract address from owner address
func (ec *Client) GetValidatorContractFromOwner(ctx context.Context, valAddr common.Address) (common.Address, error) {
	payload, err := ec.stakingUtil.Abi.Pack("valOf", valAddr)
	if err != nil {
		ec.lgr.Error("Error packing owner of validator SMC payload: ", zap.Error(err))
		return common.Address{}, err
	}
	res, err := ec.KardiaCall(ctx, ec.contructCallArgs(ec.stakingUtil.ContractAddress.Hex(), payload))
	if err != nil {
		ec.lgr.Error("GetDelegationRewards KardiaCall error: ", zap.Error(err))
		return common.Address{}, err
	}
	var owner struct {
		ValSmcAddr common.Address
	}
	err = ec.stakingUtil.Abi.UnpackIntoInterface(&owner, "valOf", res)
	if err != nil {
		ec.lgr.Error("Error unpacking owner of validator SMC error: ", zap.Error(err))
		return common.Address{}, err
	}
	return owner.ValSmcAddr, nil
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
	var rewards struct {
		Rewards *big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&rewards, "getDelegationRewards", res)
	if err != nil {
		ec.lgr.Error("Error unpacking delegation rewards: ", zap.Error(err))
		return nil, err
	}
	return rewards.Rewards, nil
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

	var delegation struct {
		Stake          *big.Int
		PreviousPeriod *big.Int
		Height         *big.Int
		Shares         *big.Int
		Owner          common.Address
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&delegation, "delegationByAddr", res)
	if err != nil {
		ec.lgr.Error("Error unpacking delegator's staked amount: ", zap.Error(err))
		return nil, err
	}
	return delegation.Stake, nil
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

	var udbEntry struct {
		Balances        []*big.Int
		CompletionTimes []*big.Int
	}
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&udbEntry, "getUBDEntries", res)
	if err != nil {
		ec.lgr.Error("Error unpacking UDB entry: ", zap.Error(err))
		return nil, nil, err
	}
	totalAmount := new(big.Int).SetInt64(0)
	withdrawableAmount := new(big.Int).SetInt64(0)
	now := new(big.Int).SetInt64(time.Now().Unix())
	for i, balance := range udbEntry.Balances {
		totalAmount = new(big.Int).Add(totalAmount, balance)
		if udbEntry.CompletionTimes[i].Cmp(now) == -1 {
			withdrawableAmount = new(big.Int).Add(withdrawableAmount, balance)
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

func (ec *Client) contructCallArgs(address string, payload []byte) types.CallArgsJSON {
	return types.CallArgsJSON{
		From:     address,
		To:       &address,
		Gas:      100000000,
		GasPrice: big.NewInt(0),
		Value:    big.NewInt(0),
		Data:     common.Bytes(payload).String(),
	}
}
