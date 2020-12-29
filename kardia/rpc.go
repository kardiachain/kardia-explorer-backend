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

	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/go-kardia/rpc"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
)

var (
	tenPoweredBy5  = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
	tenPoweredBy18 = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

type RPCClient struct {
	c      *rpc.Client
	isDead bool
	ip     string
}

type SmcUtil struct {
	Abi             *abi.ABI
	ContractAddress common.Address
	Bytecode        string
}

// Client return an *rpc.Client instance
type Client struct {
	clientList        []*RPCClient
	trustedClientList []*RPCClient
	defaultClient     *RPCClient
	numRequest        int

	stakingUtil   *SmcUtil
	validatorUtil *SmcUtil
	paramsUtil    *SmcUtil

	lgr *zap.Logger
}

// NewKaiClient creates a client that uses the given RPC client.
func NewKaiClient(config *Config) (ClientInterface, error) {
	if len(config.rpcURL) == 0 && len(config.trustedNodeRPCURL) == 0 {
		return nil, errors.New("empty RPC URL")
	}

	var (
		defaultClient     *RPCClient = nil
		clientList        []*RPCClient
		trustedClientList []*RPCClient
	)
	for _, u := range config.rpcURL {
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
	for _, u := range config.trustedNodeRPCURL {
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
		config.lgr.Error("Error reading staking contract abi", zap.Error(err))
		return nil, err
	}
	stakingUtil := &SmcUtil{
		Abi:             &stakingSmcABI,
		ContractAddress: common.HexToAddress(cfg.StakingContractAddr),
		Bytecode:        cfg.StakingContractByteCode,
	}
	validatorABI, err := os.Open(path.Join(path.Dir(filename), "../kardia/abi/validator.json"))
	if err != nil {
		panic("cannot read validator ABI file")
	}
	validatorSmcAbi, err := abi.JSON(validatorABI)
	if err != nil {
		config.lgr.Error("Error reading validator contract abi", zap.Error(err))
		return nil, err
	}
	validatorUtil := &SmcUtil{
		Abi:      &validatorSmcAbi,
		Bytecode: cfg.ValidatorContractByteCode,
	}
	paramsSmcAddr, err := GetParamsSMCAddress(stakingUtil, defaultClient)
	if err != nil {
		config.lgr.Error("Error getting params contract address", zap.Error(err))
		return nil, err
	}
	paramsABI, err := os.Open(path.Join(path.Dir(filename), "../kardia/abi/params.json"))
	if err != nil {
		panic("cannot read params ABI file")
	}
	paramsSmcAbi, err := abi.JSON(paramsABI)
	if err != nil {
		config.lgr.Error("Error reading params contract abi", zap.Error(err))
		return nil, err
	}
	paramsUtil := &SmcUtil{
		Abi:             &paramsSmcAbi,
		ContractAddress: paramsSmcAddr,
		Bytecode:        cfg.ValidatorContractByteCode,
	}

	return &Client{clientList, trustedClientList, defaultClient, 0, stakingUtil, validatorUtil, paramsUtil, config.lgr}, nil
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
func (ec *Client) BlockByHeight(ctx context.Context, height uint64) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByNumber", height)
}

// BlockHeaderByNumber returns a block header from the current canonical chain.
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

// BalanceAt returns balance (in HYDRO) of the given account.
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
	err := ec.chooseClient().c.CallContext(ctx, &result, "account_getStorageAt", common.HexToAddress(account), key, "latest")
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) GetCode(ctx context.Context, account string) (common.Bytes, error) {
	var result common.Bytes
	err := ec.chooseClient().c.CallContext(ctx, &result, "account_getCode", common.HexToAddress(account), "latest")
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

func (ec *Client) NodesInfo(ctx context.Context) ([]*types.NodeInfo, error) {
	var (
		nodes []*types.NodeInfo
		err   error
	)
	clientList := append(ec.clientList, ec.trustedClientList...)
	nodeMap := make(map[string]*types.NodeInfo, len(clientList)) // list all nodes in network
	peersMap := make(map[string]*types.RPCPeerInfo)              // list all peers details
	for _, client := range clientList {
		var (
			node  *types.NodeInfo
			peers []*types.RPCPeerInfo
		)
		// get current node info then get it's peers
		err = client.c.CallContext(ctx, &node, "node_nodeInfo")
		if err != nil {
			continue
		}
		err := client.c.CallContext(ctx, &peers, "node_peers")
		if err != nil {
			continue
		}
		node.Peers = make(map[string]*types.PeerInfo)
		for _, peer := range peers {
			// append current peer to this node
			node.Peers[peer.NodeInfo.ID] = &types.PeerInfo{
				Moniker:  peer.NodeInfo.Moniker,
				Duration: peer.ConnectionStatus.Duration,
				RemoteIP: peer.RemoteIP,
			}
			peersMap[peer.NodeInfo.ID] = peer
			// try to discover new nodes through these peers
			// init a new node info in list
			if nodeMap[peer.NodeInfo.ID] == nil {
				nodeMap[peer.NodeInfo.ID] = peer.NodeInfo
			}
			if nodeMap[peer.NodeInfo.ID].Peers == nil {
				nodeMap[peer.NodeInfo.ID].Peers = make(map[string]*types.PeerInfo)
			}
			nodeMap[peer.NodeInfo.ID].Peers[node.ID] = &types.PeerInfo{} // mark this peer for re-updating later
		}
		nodeMap[node.ID] = node
	}
	for _, node := range nodeMap {
		// re-update full peers info
		for id, peer := range node.Peers {
			node.Peers[id] = &types.PeerInfo{
				Duration: peer.Duration,
				Moniker:  peer.Moniker,
				RemoteIP: peer.RemoteIP,
			}
		}
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
	valsSet, err := ec.GetValidatorSets(ctx)
	if err != nil {
		return nil, err
	}
	// update validator's role
	validator.Role = ec.getValidatorRole(valsSet, validator.Address, validator.Status)
	// calculate his rate from big.Int
	convertedVal, err := convertValidatorInfo(validator, nil, validator.Role)
	if err != nil {
		return nil, err
	}
	return convertedVal, nil
}

//Validators return list validators from network
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
			delegators[del.Address.Hex()] = true
			// exclude validator self delegation
			if del.Address.Equal(val.Address) {
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
		val.Role = ec.getValidatorRole(valsSet, val.Address, val.Status)
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
	result := &types.Validators{
		TotalValidators:            totalValidators,
		TotalDelegators:            len(delegators),
		TotalProposers:             totalProposers,
		TotalCandidates:            totalCandidates,
		TotalStakedAmount:          totalStakedAmount.String(),
		TotalValidatorStakedAmount: new(big.Int).Sub(totalStakedAmount, totalDelegatorStakedAmount).String(),
		TotalDelegatorStakedAmount: totalDelegatorStakedAmount.String(),
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
