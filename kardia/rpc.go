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
	"github.com/kardiachain/go-kardiamain/lib/rlp"
	"github.com/kardiachain/go-kardiamain/rpc"
	coreTypes "github.com/kardiachain/go-kardiamain/types"
)

// RPCClient return an *rpc.Client instance
type Client struct {
	c   *rpc.Client
	lgr *zap.Logger
}

// NewKaiClient creates a client that uses the given RPC client.
func NewKaiClient(rpcUrl string, Lgr *zap.Logger) (*Client, error) {
	rpcClient, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial rpc %q: %v", rpcUrl, err)
	}
	return &Client{rpcClient, Lgr}, nil
}

// LatestBlockNumber gets latest block number
func (ec *Client) LatestBlockNumber(ctx context.Context) (uint64, error) {
	var result uint64
	err := ec.c.CallContext(ctx, &result, "kai_blockNumber")
	return result, err
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByHash", hash)
}

// BlockByNumber returns a block from the current canonical chain.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
// TODO(trinhdn): If number is nil, the latest known block is returned.
func (ec *Client) BlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByNumber", number)
}

// BlockHeaderByNumber returns a block header from the current canonical chain.
// TODO(trinhdn): If number is nil, the latest known block header is returned.
func (ec *Client) BlockHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error) {
	return ec.getBlockHeader(ctx, "kai_getBlockHeaderByNumber", number)
}

// BlockHeaderByHash returns the given block header.
func (ec *Client) BlockHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return ec.getBlockHeader(ctx, "kai_getBlockHeaderByHash", hash)
}

// TransactionByHash returns the transaction with the given hash.
func (ec *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	var raw *types.Transaction
	err = ec.c.CallContext(ctx, &raw, "tx_getTransaction", hash)
	if err != nil {
		return nil, false, err
	} else if raw == nil {
		return nil, false, kardia.NotFound
	}
	// if raw.From != "" && raw.BlockHash != "" {
	// 	setSenderFromServer(ctx, raw, common.HexToAddress(raw.From), common.HexToHash(raw.BlockHash))
	// }
	return raw, raw.BlockNumber == 0, nil
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockHash common.Hash, blockNumber uint64) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "account_balance", account, blockHash, blockNumber)
	return result, err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber uint64) ([]byte, error) {
	var result common.Bytes
	err := ec.c.CallContext(ctx, &result, "kai_getStorageAt", account, key, blockNumber)
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account common.Address, blockNumber uint64) ([]byte, error) {
	var result common.Bytes
	err := ec.c.CallContext(ctx, &result, "kai_getCode", account, blockNumber)
	return result, err
}

// NonceAt returns the account nonce of the given account.
func (ec *Client) NonceAt(ctx context.Context, account common.Address) (uint64, error) {
	var result uint64
	err := ec.c.CallContext(ctx, &result, "account_nonce", account)
	return result, err
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
// TODO(trinhdn): verify which types of tx is suitable for this API
func (ec *Client) SendRawTransaction(ctx context.Context, tx *coreTypes.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.c.CallContext(ctx, nil, "tx_sendRawTransaction", common.ToHex(data))
}

func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	var raw types.Block
	err := ec.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, fmt.Errorf("marshal block failed: %v", err)
	}
	raw.NonceBool = false
	// get parent block for its hash
	if num, ok := args[0].(uint64); ok {
		if num == 0 {
			return &raw, nil
		}
		var rawParentBlockHeader types.Header
		err = ec.c.CallContext(ctx, &rawParentBlockHeader, "kai_getBlockHeaderByNumber", num-1)
		if err != nil {
			return nil, fmt.Errorf("marshal parent block header failed: %v", err)
		}
		raw.ParentHash = rawParentBlockHeader.Hash
	}
	return &raw, nil
}

func (ec *Client) getBlockHeader(ctx context.Context, method string, args ...interface{}) (*types.Header, error) {
	var raw types.Header
	err := ec.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, fmt.Errorf("marshal block header failed: %v", err)
	}
	return &raw, nil
}
