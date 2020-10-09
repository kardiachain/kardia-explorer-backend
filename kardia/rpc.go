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
	"fmt"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
	kardia "github.com/kardiachain/go-kardiamain"
	"github.com/kardiachain/go-kardiamain/lib/common"
	"github.com/kardiachain/go-kardiamain/lib/p2p"
	"github.com/kardiachain/go-kardiamain/lib/rlp"
	"github.com/kardiachain/go-kardiamain/rpc"
	coreTypes "github.com/kardiachain/go-kardiamain/types"
)

type RPCClient struct {
	c      *rpc.Client
	isDead bool
	ip     string
}

// Client return an *rpc.Client instance
type Client struct {
	clientList    []*RPCClient
	defaultClient *RPCClient
	numRequest    uint64
	lgr           *zap.Logger
}

type Validator struct {
	address     string
	votingPower float64
}

// NewKaiClient creates a client that uses the given RPC client.
func NewKaiClient(rpcURL []string, lgr *zap.Logger) (ClientInterface, error) {
	if len(rpcURL) == 0 {
		return nil, fmt.Errorf("Empty RPC URL")
	}
	var clientList = []*RPCClient{}
	for _, u := range rpcURL {
		rpcClient, err := rpc.Dial(u)
		if err != nil {
			return nil, fmt.Errorf("Failed to dial rpc %q: %v", u, err)
		}
		clientList = append(clientList, &RPCClient{
			c:      rpcClient,
			isDead: false,
			ip:     u,
		})
	}

	return &Client{clientList, clientList[len(clientList)-1], 0, lgr}, nil
}

func (ec *Client) chooseClient() *RPCClient {
	if len(ec.clientList) > 1 {
		if ec.numRequest >= 50000 {
			ec.numRequest = 0
		}
		return ec.clientList[ec.numRequest%(uint64(len(ec.clientList))-1)]
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
func (ec *Client) GetTransaction(ctx context.Context, hash string) (*types.Transaction, bool, error) {
	var raw *types.Transaction
	err := ec.defaultClient.c.CallContext(ctx, &raw, "tx_getTransaction", common.HexToHash(hash))
	if err != nil {
		return nil, false, err
	} else if raw == nil {
		return nil, false, kardia.NotFound
	}
	return raw, raw.BlockNumber == 0, nil
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	RPCClient := ec.chooseClient()
	var r *types.Receipt
	err := RPCClient.c.CallContext(ctx, &r, "tx_getTransactionReceipt", common.HexToHash(txHash))
	if err == nil {
		if r == nil {
			return nil, kardia.NotFound
		}
	}
	return r, err
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account string, args interface{}) (string, error) {
	var result string
	var err error
	if blockHeight, ok := args.(uint64); ok {
		err = ec.defaultClient.c.CallContext(ctx, &result, "account_balance", common.HexToAddress(account), nil, blockHeight)
	} else if blockHash, ok := args.(string); ok {
		err = ec.defaultClient.c.CallContext(ctx, &result, "account_balance", common.HexToAddress(account), blockHash, nil)
	}
	return result, err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StorageAt(ctx context.Context, account string, key string, blockNumber uint64) ([]byte, error) {
	var result common.Bytes
	err := ec.defaultClient.c.CallContext(ctx, &result, "kai_getStorageAt", common.HexToAddress(account), key, blockNumber)
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account string, blockNumber uint64) ([]byte, error) {
	var result common.Bytes
	err := ec.defaultClient.c.CallContext(ctx, &result, "kai_getCode", common.HexToAddress(account), blockNumber)
	return result, err
}

// NonceAt returns the account nonce of the given account.
func (ec *Client) NonceAt(ctx context.Context, account string) (uint64, error) {
	var result uint64
	err := ec.defaultClient.c.CallContext(ctx, &result, "account_nonce", common.HexToAddress(account))
	return result, err
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the GetTransactionReceipt method to get the
// contract address after the transaction has been mined.
// TODO(trinhdn): verify which types of tx is suitable for this API
func (ec *Client) SendRawTransaction(ctx context.Context, tx *coreTypes.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.defaultClient.c.CallContext(ctx, nil, "tx_sendRawTransaction", common.ToHex(data))
}

func (ec *Client) Peers(ctx context.Context) ([]*p2p.PeerInfo, error) {
	var result []*p2p.PeerInfo
	err := ec.defaultClient.c.CallContext(ctx, &result, "node_peers")
	return result, err
}

func (ec *Client) NodeInfo(ctx context.Context) (*p2p.NodeInfo, error) {
	var result *p2p.NodeInfo
	err := ec.defaultClient.c.CallContext(ctx, &result, "node_nodeInfo")
	return result, err
}

func (ec *Client) Datadir(ctx context.Context) (string, error) {
	var result string
	err := ec.defaultClient.c.CallContext(ctx, &result, "node_datadir")
	return result, err
}

func (ec *Client) Validator(ctx context.Context) *Validator {
	var result map[string]interface{}
	_ = ec.defaultClient.c.CallContext(ctx, &result, "kai_validator")
	var ret = &Validator{}
	for key, value := range result {
		if key == "address" {
			ret.address = value.(string)
		} else if key == "votingPower" {
			ret.votingPower = value.(float64)
		}
	}
	return ret
}

func (ec *Client) Validators(ctx context.Context) []*Validator {
	var result []map[string]interface{}
	_ = ec.defaultClient.c.CallContext(ctx, &result, "kai_validators")
	var ret []*Validator
	for _, val := range result {
		var tmp = &Validator{}
		for key, value := range val {
			if key == "address" {
				tmp.address = value.(string)
			} else if key == "votingPower" {
				tmp.votingPower = value.(float64)
			}
		}
		ret = append(ret, tmp)
	}
	return ret
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
