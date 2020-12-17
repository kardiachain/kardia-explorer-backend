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

	stakingUtil     *staking.StakingSmcUtil
	validatorUtil   *staking.ValidatorSmcUtil
	totalValidators int

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
		panic("cannot read validator ABI file")
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

	return &Client{clientList, trustedClientList, defaultClient, 0, stakingUtil, validatorUtil, cfg.totalValidators, cfg.lgr}, nil
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
	validator, err := ec.getValidatorFromSMC(ctx, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	// update validator's role. If he's in validators set, he is a proposer
	valsSet, err := ec.GetValidatorSets(ctx)
	if err != nil {
		return nil, err
	}
	for _, val := range valsSet {
		if val.Equal(common.HexToAddress(address)) {
			validator.Status = 2
			return convertValidator(validator), nil
		}
	}
	// else if his staked amount is enough, he is a normal validator
	minStakedAmount, ok := new(big.Int).SetString(cfg.MinStakedAmount, 10)
	if !ok {
		ec.lgr.Error("error parsing MinStakedAmount to big.Int:", zap.String("MinStakedAmount", cfg.MinStakedAmount), zap.Any("value", minStakedAmount))
	}
	if validator.Tokens.Cmp(minStakedAmount) >= 0 {
		validator.Status = 1
		return convertValidator(validator), nil
	}
	// otherwise he is a registered validator
	validator.Status = 0
	return convertValidator(validator), nil
}

func (ec *Client) Validators(ctx context.Context) (*types.Validators, error) {
	var (
		proposersStakedAmount = big.NewInt(0)
		validators            []*types.RPCValidator
	)
	validators, err := ec.getValidatorsFromSMC(ctx)
	if err != nil {
		return nil, err
	}
	// compare staked amount btw validators to determine their status
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Tokens.Cmp(validators[j].Tokens) == 1
	})
	var (
		delegators                 = make(map[string]bool)
		totalProposers             = 0
		totalStakedAmount          = big.NewInt(0)
		totalDelegatorStakedAmount = big.NewInt(0)

		ok bool
	)
	minStakedAmount, ok := new(big.Int).SetString(cfg.MinStakedAmount, 10)
	if !ok {
		ec.lgr.Error("error parsing MinStakedAmount to big.Int:", zap.String("MinStakedAmount", cfg.MinStakedAmount), zap.Any("value", minStakedAmount))
	}
	for i, val := range validators {
		for _, del := range val.Delegators {
			delegators[del.Address.Hex()] = true
			// exclude validator self delegation
			if del.Address.Equal(val.ValAddr) {
				continue
			}
			totalDelegatorStakedAmount = new(big.Int).Add(totalDelegatorStakedAmount, del.StakedAmount)
		}
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, val.Tokens)
		if validators[i].Tokens.Cmp(minStakedAmount) == -1 || val.Status < 2 {
			val.Status = 0 // validator who has staked under 12.5M KAI is considers a registered one
		} else if totalProposers < ec.totalValidators {
			val.Status = 2 // validator who has staked over 12.5M KAI and belong to top 20 of validator based on voting power is considered a proposer
			totalProposers++
			proposersStakedAmount = new(big.Int).Add(proposersStakedAmount, validators[i].Tokens)
		} else {
			val.Status = 1 // validator who has staked over 12.5M KAI and not belong to top 20 of validator based on voting power is considered a normal validator
		}
	}
	var returnValsList []*types.Validator
	for _, val := range validators {
		convertedVal, err := convertValidatorInfo(val, proposersStakedAmount, val.Status)
		if err != nil {
			return nil, err
		}
		returnValsList = append(returnValsList, convertedVal)
	}
	result := &types.Validators{
		TotalValidators:            len(validators),
		TotalDelegators:            len(delegators),
		TotalStakedAmount:          totalStakedAmount.String(),
		TotalValidatorStakedAmount: new(big.Int).Sub(totalStakedAmount, totalDelegatorStakedAmount).String(),
		TotalDelegatorStakedAmount: totalDelegatorStakedAmount.String(),
		TotalProposer:              totalProposers,
		Validators:                 returnValsList,
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

func convertValidatorInfo(srcVal *types.RPCValidator, totalStakedAmount *big.Int, status uint8) (*types.Validator, error) {
	var err error
	val := convertValidator(srcVal)
	if val.CommissionRate, err = convertBigIntToPercentage(srcVal.CommissionRate); err != nil {
		return nil, err
	}
	if val.MaxRate, err = convertBigIntToPercentage(srcVal.MaxRate); err != nil {
		return nil, err
	}
	if val.MaxChangeRate, err = convertBigIntToPercentage(srcVal.MaxChangeRate); err != nil {
		return nil, err
	}
	if totalStakedAmount != nil && status == 2 {
		if val.VotingPowerPercentage, err = calculateVotingPower(srcVal.Tokens, totalStakedAmount); err != nil {
			return nil, err
		}
	} else {
		val.VotingPowerPercentage = "0"
	}
	return val, nil
}

func convertBigIntToPercentage(input *big.Int) (string, error) {
	tmp := new(big.Int).Mul(input, tenPoweredBy18)
	result := new(big.Int).Div(tmp, tenPoweredBy18).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-16]+"."+result[len(result)-16:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func calculateVotingPower(valStakedAmount *big.Int, total *big.Int) (string, error) {
	tmp := new(big.Int).Mul(valStakedAmount, tenPoweredBy5)
	result := new(big.Int).Div(tmp, total).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-3]+"."+result[len(result)-3:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func (ec *Client) getValidatorsFromSMC(ctx context.Context) ([]*types.RPCValidator, error) {
	allValsLen, err := ec.GetAllValsLength(ctx)
	if err != nil {
		return nil, err
	}
	var (
		one      = big.NewInt(1)
		valsInfo []*types.RPCValidator
	)
	for i := new(big.Int).SetInt64(0); i.Cmp(allValsLen) < 0; i.Add(i, one) {
		valContractAddr, err := ec.GetValSmcAddr(ctx, i)
		if err != nil {
			return nil, err
		}
		valInfo, err := ec.GetValidatorInfo(ctx, valContractAddr)
		if err != nil {
			return nil, err
		}
		valInfo.Delegators, err = ec.GetDelegators(ctx, valContractAddr)
		if err != nil {
			return nil, err
		}
		valInfo.ValStakingSmc = valContractAddr
		valsInfo = append(valsInfo, valInfo)
	}
	return valsInfo, nil
}

func (ec *Client) getValidatorFromSMC(ctx context.Context, valAddr common.Address) (*types.RPCValidator, error) {
	valContractAddr, err := ec.GetValFromOwner(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	val, err := ec.GetValidatorInfo(ctx, valContractAddr)
	if err != nil {
		return nil, err
	}
	val.Delegators, err = ec.GetDelegators(ctx, valContractAddr)
	if err != nil {
		return nil, err
	}
	val.ValStakingSmc = valContractAddr
	return val, nil
}

func convertValidator(src *types.RPCValidator) *types.Validator {
	var name []byte
	for _, b := range src.Name {
		if b != 0 {
			name = append(name, byte(b))
		}
	}
	var delegators []*types.Delegator
	for _, del := range src.Delegators {
		delegators = append(delegators, &types.Delegator{
			Address:      del.Address,
			StakedAmount: del.StakedAmount.String(),
			Reward:       del.Reward.String(),
		})
	}
	return &types.Validator{
		Address:               src.ValAddr,
		SmcAddress:            src.ValStakingSmc,
		Status:                src.Status,
		Jailed:                src.Jailed,
		Name:                  string(name),
		VotingPowerPercentage: "",
		StakedAmount:          src.Tokens.String(),
		AccumulatedCommission: src.AccumulatedCommission.String(),
		CommissionRate:        "",
		TotalDelegators:       len(src.Delegators),
		MaxRate:               "",
		MaxChangeRate:         "",
		Delegators:            delegators,
	}
}
