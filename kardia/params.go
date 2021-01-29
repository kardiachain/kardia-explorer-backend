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

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/go-kardia/lib/common"
)

// GetMaxProposers returns max number of proposers
func (ec *Client) GetMaxProposers(ctx context.Context) (int64, error) {
	payload, err := ec.paramsUtil.Abi.Pack("getMaxProposers")
	if err != nil {
		return 0, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return 0, err
	}

	var result struct {
		MaxProposers *big.Int
	}
	// unpack result
	err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "getMaxProposers", res)
	if err != nil {
		ec.lgr.Error("Error unpacking max proposers", zap.Error(err))
		return 0, err
	}
	return result.MaxProposers.Int64(), nil
}

// GetParams returns value of params by indexes
func (ec *Client) GetParams(ctx context.Context) (map[string]*big.Int, error) {
	params := make(map[string]*big.Int)
	for i, paramName := range cfg.ParamKeys {
		payload, err := ec.paramsUtil.Abi.Pack("getParam", uint8(i))
		if err != nil {
			return nil, err
		}
		res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
		if err != nil {
			return nil, err
		}

		var result struct {
			Value *big.Int
		}
		// unpack result
		err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "getParam", res)
		if err != nil {
			ec.lgr.Error("Error unpacking params", zap.Int("id", i), zap.String("name", paramName), zap.Error(err))
			return nil, err
		}
		params[cfg.ParamKeys[i]] = result.Value
	}
	return params, nil
}

// GetProposalDetails returns detail of a proposal by ID
func (ec *Client) GetProposalDetails(ctx context.Context, proposalId *big.Int) (*types.ProposalDetail, error) {
	payload, err := ec.paramsUtil.Abi.Pack("getProposalDetails", proposalId)
	if err != nil {
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return nil, err
	}

	var result struct {
		VoteYes     *big.Int
		VoteNo      *big.Int
		VoteAbstain *big.Int
		ParamKeys   []uint8
		ParamValues []*big.Int
	}
	// unpack result
	err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "getProposalDetails", res)
	if err != nil {
		ec.lgr.Error("Error unpacking proposal", zap.String("ID", proposalId.String()), zap.Error(err))
		return nil, err
	}
	metadata, err := ec.getProposalMetadata(ctx, proposalId)
	if err != nil {
		return nil, err
	}
	detail := &types.ProposalDetail{
		ProposalMetadata: *metadata,
		VoteYes:          result.VoteYes.Uint64(),
		VoteNo:           result.VoteNo.Uint64(),
		VoteAbstain:      result.VoteAbstain.Uint64(),
		Params:           make(map[string]string),
	}
	for i, key := range result.ParamKeys {
		detail.Params[cfg.ParamKeys[key]] = result.ParamValues[i].String()
	}
	return detail, nil
}

// GetProposalMetadata returns metadata of a proposal by ID
func (ec *Client) getProposalMetadata(ctx context.Context, proposalId *big.Int) (*types.ProposalMetadata, error) {
	payload, err := ec.paramsUtil.Abi.Pack("proposals", proposalId)
	if err != nil {
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return nil, err
	}

	var result struct {
		Proposer  common.Address
		StartTime *big.Int
		EndTime   *big.Int
		Deposit   *big.Int
		Status    uint8
	}
	// unpack result
	err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "proposals", res)
	if err != nil {
		ec.lgr.Error("Error unpacking proposal metadata", zap.String("ID", proposalId.String()), zap.Error(err))
		return nil, err
	}
	return &types.ProposalMetadata{
		ID:        proposalId.Uint64(),
		Proposer:  result.Proposer.String(),
		StartTime: result.StartTime.Uint64(),
		EndTime:   result.EndTime.Uint64(),
		Deposit:   result.Deposit.String(),
		Status:    result.Status,
	}, nil
}

// GetTotalProposals returns total number of proposals
func (ec *Client) GetTotalProposals(ctx context.Context) (*big.Int, error) {
	payload, err := ec.paramsUtil.Abi.Pack("allProposal")
	if err != nil {
		return nil, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return nil, err
	}

	var result struct {
		Total *big.Int
	}
	// unpack result
	err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "allProposal", res)
	if err != nil {
		ec.lgr.Error("Error unpacking total proposals", zap.Error(err))
		return nil, err
	}
	return result.Total, nil
}

// GetProposals returns list of proposals
func (ec *Client) GetProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalMetadata, uint64, error) {
	one := big.NewInt(1)
	start := new(big.Int).SetInt64(int64(pagination.Skip))
	end := new(big.Int).SetInt64(int64(pagination.Limit))
	total, err := ec.GetTotalProposals(ctx)
	if err != nil {
		return nil, 0 , err
	}
	if end.Cmp(total) == 1 {
		end = total
	}
	var result []*types.ProposalMetadata
	// i must be a new int so that it does not overwrite start
	for i := new(big.Int).Set(start); i.Cmp(end) < 0; i.Add(i, one) {
		metadata, err := ec.getProposalMetadata(ctx, i)
		if err != nil {
			continue
		}
		result = append(result, metadata)
	}
	return result, total.Uint64(), nil
}
