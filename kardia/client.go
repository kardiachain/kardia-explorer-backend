/*
 *  Copyright 2018 KardiaChain
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
// Package kardia
package kardia

import (
	"context"
	"math/big"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

type ClientInterface interface {
	LatestBlockNumber(ctx context.Context) (uint64, error)
	BlockByHash(ctx context.Context, hash string) (*types.Block, error)
	BlockByHeight(ctx context.Context, height uint64) (*types.Block, error)
	GetTransaction(ctx context.Context, hash string) (*types.Transaction, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error)
	GetBalance(ctx context.Context, account string) (string, error)
	GetCode(ctx context.Context, account string) (common.Bytes, error)
	NodesInfo(ctx context.Context) ([]*types.NodeInfo, error)
	Validator(ctx context.Context, address string) (*types.Validator, error)
	Validators(ctx context.Context) ([]*types.Validator, error)

	// staking related methods
	GetValidatorsByDelegator(ctx context.Context, delAddr common.Address) ([]*types.ValidatorsByDelegator, error)
	GetTotalSlashedToken(ctx context.Context) (*big.Int, error)
	GetCirculatingSupply(ctx context.Context) (*big.Int, error)

	// validator related methods
	GetSlashEvents(ctx context.Context, valAddr common.Address) ([]*types.SlashEvents, error)

	// params related methods
	GetMaxProposers(ctx context.Context) (int64, error)
	GetParams(ctx context.Context) ([]*types.NetworkParams, error)
	GetProposalDetails(ctx context.Context, proposalID *big.Int) (*types.ProposalDetail, error)
	GetProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalDetail, uint64, error)

	// utilities methods
	DecodeInputData(to string, input string) (*types.FunctionCall, error)
	NonceAt(ctx context.Context, account string) (uint64, error)
	KardiaCall(ctx context.Context, args types.CallArgsJSON) (common.Bytes, error)
	DecodeInputWithABI(to string, input string, smcABI *abi.ABI) (*types.FunctionCall, error)
}

type Config struct {
	rpcURL            []string
	trustedNodeRPCURL []string
	lgr               *zap.Logger
}

func NewConfig(rpcURL []string, trustedNodeRPCURL []string, lgr *zap.Logger) *Config {
	return &Config{
		rpcURL:            rpcURL,
		trustedNodeRPCURL: trustedNodeRPCURL,
		lgr:               lgr,
	}
}
