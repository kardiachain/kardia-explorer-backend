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
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/kardiachain/go-kardia/configs"
	"github.com/labstack/gommon/log"

	"go.uber.org/zap"

	kardia "github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/go-kardia/mainchain/staking"
	"github.com/kardiachain/go-kardia/rpc"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
)

var (
	ErrParsingBigIntFromString = errors.New("cannot parse string to big.Int")
	ErrValidatorNotFound       = errors.New("validator address not found")

	tenPoweredBy5  = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
	tenPoweredBy18 = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

type RPCClient struct {
	c      *rpc.Client
	isDead bool
	ip     string
}

// Client return an *rpc.Client instance
type Client struct {
	clientList        []*RPCClient
	trustedClientList []*RPCClient
	defaultClient     *RPCClient
	numRequest        int

	stakingUtil        *staking.StakingSmcUtil
	validatorUtil      *staking.ValidatorSmcUtil
	maxTotalValidators int

	lgr *zap.Logger
}

// NewKaiClient creates a client that uses the given RPC client.
func NewKaiClient(cfg *Config) (ClientInterface, error) {
	if len(cfg.rpcURL) == 0 && len(cfg.trustedNodeRPCURL) == 0 {
		return nil, errors.New("empty RPC URL")
	}

	var (
		defaultClient *RPCClient = nil
		clientList               = []*RPCClient{}
	)
	for _, u := range cfg.rpcURL {
		rpcClient, err := rpc.Dial(u)
		if err != nil {
			return nil, err
		}
		newClient := &RPCClient{
			c:      rpcClient,
			isDead: false,
			ip:     u,
		}
		clientList = append(clientList, newClient)
	}
	var trustedClientList = []*RPCClient{}
	for _, u := range cfg.trustedNodeRPCURL {
		rpcClient, err := rpc.Dial(u)
		if err != nil {
			return nil, err
		}
		newClient := &RPCClient{
			c:      rpcClient,
			isDead: false,
			ip:     u,
		}
		trustedClientList = append(trustedClientList, newClient)
	}
	// set default RPC client as one of our trusted ones
	defaultClient = trustedClientList[0]

	_, filename, _, _ := runtime.Caller(1)
	stakingABI, err := os.Open(path.Join(path.Dir(filename), "../kardia/abi/staking.json"))
	if err != nil {
		panic("cannot read staking ABI file")
	}
	stakingSmcABI, err := abi.JSON(stakingABI)
	if err != nil {
		log.Error("Error reading staking contract abi", "err", err)
		return nil, err
	}
	stakingUtil := &staking.StakingSmcUtil{
		Abi:             &stakingSmcABI,
		ContractAddress: common.HexToAddress(configs.StakingContract.Address),
		Bytecode:        configs.StakingContract.ByteCode,
	}
	validatorABI, err := os.Open(path.Join(path.Dir(filename), "../kardia/abi/validator.json"))
	if err != nil {
		panic("cannot read staking ABI file")
	}
	validatorSmcAbi, err := abi.JSON(validatorABI)
	if err != nil {
		log.Error("Error reading validator contract abi", "err", err)
		return nil, err
	}
	validatorUtil := &staking.ValidatorSmcUtil{
		Abi:      &validatorSmcAbi,
		Bytecode: configs.ValidatorContract.ByteCode,
	}

	return &Client{clientList, trustedClientList, defaultClient, 0, stakingUtil, validatorUtil, cfg.maxTotalValidators, cfg.lgr}, nil
}

func (ec *Client) chooseClient() *RPCClient {
	if len(ec.clientList) > 1 {
		if ec.numRequest == len(ec.clientList)-1 {
			ec.numRequest = 0
		} else {
			ec.numRequest++
		}
		return ec.clientList[ec.numRequest%(len(ec.clientList)-1)]
	}
	return ec.defaultClient
}

// LatestBlockNumber gets latest block number
func (ec *Client) LatestBlockNumber(ctx context.Context) (uint64, error) {
	var result uint64
	err := ec.defaultClient.c.CallContext(ctx, &result, "kai_blockNumber")
	return result, err
}

// BlockByHash returns the given full block.
//
// Use HeaderByHash if you don't need all transactions or uncle headers.
func (ec *Client) BlockByHash(ctx context.Context, hash string) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByHash", common.HexToHash(hash))
}

// BlockByHeight returns a block from the current canonical chain.
//
// Use HeaderByNumber if you don't need all transactions or uncle headers.
// TODO(trinhdn): If number is nil, the latest known block is returned.
func (ec *Client) BlockByHeight(ctx context.Context, height uint64) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByNumber", height)
}

// BlockHeaderByNumber returns a block header from the current canonical chain.
// TODO(trinhdn): If number is nil, the latest known block header is returned.
func (ec *Client) BlockHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error) {
	return ec.getBlockHeader(ctx, "kai_getBlockHeaderByNumber", number)
}

// BlockHeaderByHash returns the given block header.
func (ec *Client) BlockHeaderByHash(ctx context.Context, hash string) (*types.Header, error) {
	return ec.getBlockHeader(ctx, "kai_getBlockHeaderByHash", common.HexToHash(hash))
}

// GetTransaction returns the transaction with the given hash.
func (ec *Client) GetTransaction(ctx context.Context, hash string) (*types.Transaction, error) {
	var raw *types.Transaction
	err := ec.chooseClient().c.CallContext(ctx, &raw, "tx_getTransaction", common.HexToHash(hash))
	if err != nil {
		return nil, err
	} else if raw == nil {
		return nil, kardia.NotFound
	}
	return raw, nil
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	var r *types.Receipt
	err := ec.chooseClient().c.CallContext(ctx, &r, "tx_getTransactionReceipt", common.HexToHash(txHash))
	if err == nil {
		if r == nil {
			return nil, kardia.NotFound
		}
	}
	return r, err
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) GetBalance(ctx context.Context, account string) (string, error) {
	var (
		result string
		err    error
	)
	err = ec.chooseClient().c.CallContext(ctx, &result, "account_balance", common.HexToAddress(account), "latest")
	return result, err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) GetStorageAt(ctx context.Context, account string, key string) (common.Bytes, error) {
	var result common.Bytes
	err := ec.chooseClient().c.CallContext(ctx, &result, "kai_getStorageAt", common.HexToAddress(account), key, "latest")
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) GetCode(ctx context.Context, account string) (common.Bytes, error) {
	var result common.Bytes
	err := ec.chooseClient().c.CallContext(ctx, &result, "kai_getCode", common.HexToAddress(account), "latest")
	return result, err
}

// NonceAt returns the account nonce of the given account.
func (ec *Client) NonceAt(ctx context.Context, account string) (uint64, error) {
	var result uint64
	err := ec.chooseClient().c.CallContext(ctx, &result, "account_nonce", common.HexToAddress(account))
	return result, err
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the GetTransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendRawTransaction(ctx context.Context, tx string) error {
	return ec.chooseClient().c.CallContext(ctx, nil, "tx_sendRawTransaction", tx)
}

func (ec *Client) KardiaCall(ctx context.Context, args types.CallArgsJSON) (common.Bytes, error) {
	var result common.Bytes
	err := ec.chooseClient().c.CallContext(ctx, &result, "kai_kardiaCall", args, "latest")
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ec *Client) Peers(ctx context.Context, client *RPCClient) ([]*types.PeerInfo, error) {
	var result []*types.PeerInfo
	err := client.c.CallContext(ctx, &result, "node_peers")
	return result, err
}

func (ec *Client) NodesInfo(ctx context.Context) ([]*types.NodeInfo, error) {
	var (
		nodes = []*types.NodeInfo(nil)
		err   error
	)
	clientList := append(ec.clientList, ec.trustedClientList...)
	nodeMap := make(map[string]*types.NodeInfo, len(clientList))
	for _, client := range clientList {
		var (
			node  *types.NodeInfo
			peers []*types.PeerInfo
		)
		err = client.c.CallContext(ctx, &node, "node_nodeInfo")
		if err != nil {
			continue
		}
		peers, err = ec.Peers(ctx, client)
		if err != nil {
			continue
		}
		node.Peers = peers
		nodeMap[node.ID] = node
	}
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (ec *Client) Datadir(ctx context.Context) (string, error) {
	var result string
	err := ec.chooseClient().c.CallContext(ctx, &result, "node_datadir")
	return result, err
}

func (ec *Client) Validator(ctx context.Context, address string) (*types.Validator, error) {
	var validator *types.Validator
	err := ec.defaultClient.c.CallContext(ctx, &validator, "kai_validator", address, true)
	if err != nil {
		return nil, err
	}
	return validator, nil
}

func (ec *Client) Validators(ctx context.Context) (*types.Validators, error) {
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
		totalStakedAmount          = big.NewInt(0)
		totalDelegatorStakedAmount = big.NewInt(0)

		valStakedAmount *big.Int
		delStakedAmount *big.Int
		ok              bool
	)
	for _, val := range validators {
		for _, del := range val.Delegators {
			delegators[del.Address.Hex()] = true
			// exclude validator self delegation
			if del.Address.Equal(val.Address) {
				continue
			}
			delStakedAmount, ok = new(big.Int).SetString(del.StakedAmount, 10)
			if !ok {
				return nil, err
			}
			totalDelegatorStakedAmount = new(big.Int).Add(totalDelegatorStakedAmount, delStakedAmount)
		}
		valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
		if !ok {
			return nil, err
		}
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, valStakedAmount)
	}
	minStakedAmount, ok := new(big.Int).SetString(cfg.MinStakedAmount, 10)
	if !ok {
		ec.lgr.Error("error parsing MinStakedAmount to big.Int:", zap.String("MinStakedAmount", cfg.MinStakedAmount), zap.Any("value", minStakedAmount))
	}
	totalProposers := 0
	for i, val := range validators {
		valInfo, err := ec.GetValidatorInfo(ctx, val.SmcAddress)
		if err != nil {
			return nil, err
		}
		val.Status = valInfo.Status
		if stakedAmount, ok := new(big.Int).SetString(validators[i].StakedAmount, 10); ok {
			if stakedAmount.Cmp(minStakedAmount) == -1 || val.Status < 2 {
				val.Role = 0 // validator who has staked under 12.5M KAI is considers a registered one
			} else if totalProposers < ec.maxTotalValidators {
				val.Role = 2 // validator who has staked over 12.5M KAI and belong to top 20 of validator based on voting power is considered a proposer
				totalProposers++
				valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
				if !ok {
					return nil, err
				}
				proposersStakedAmount = new(big.Int).Add(proposersStakedAmount, valStakedAmount)
			} else {
				val.Role = 1 // validator who has staked over 12.5M KAI and not belong to top 20 of validator based on voting power is considered a normal validator
			}
		}
		if val, err = convertValidatorInfo(val, proposersStakedAmount, val.Role); err != nil {
			return nil, err
		}
	}
	result := &types.Validators{
		TotalValidators:            len(validators),
		TotalDelegators:            len(delegators),
		TotalStakedAmount:          totalStakedAmount.String(),
		TotalValidatorStakedAmount: new(big.Int).Sub(totalStakedAmount, totalDelegatorStakedAmount).String(),
		TotalDelegatorStakedAmount: totalDelegatorStakedAmount.String(),
		TotalProposer:              totalProposers,
		Validators:                 validators,
	}
	return result, nil
}

func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	var raw types.Block
	err := ec.defaultClient.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func (ec *Client) getBlockHeader(ctx context.Context, method string, args ...interface{}) (*types.Header, error) {
	var raw types.Header
	err := ec.defaultClient.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func convertValidatorInfo(val *types.Validator, totalStakedAmount *big.Int, role int) (*types.Validator, error) {
	var err error
	if val.CommissionRate, err = convertBigIntToPercentage(val.CommissionRate); err != nil {
		return nil, err
	}
	if val.MaxRate, err = convertBigIntToPercentage(val.MaxRate); err != nil {
		return nil, err
	}
	if val.MaxChangeRate, err = convertBigIntToPercentage(val.MaxChangeRate); err != nil {
		return nil, err
	}
	if totalStakedAmount != nil && role == 2 {
		if val.VotingPowerPercentage, err = calculateVotingPower(val.StakedAmount, totalStakedAmount); err != nil {
			return nil, err
		}
	} else {
		val.VotingPowerPercentage = "0"
	}
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
