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
	"errors"
	"math/big"
	"os"
	"path"
	"runtime"

	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/go-kardia/rpc"

	"github.com/kardiachain/explorer-backend/cfg"
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
