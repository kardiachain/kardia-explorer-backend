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
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/kardiachain/explorer-backend/types"
	kardia "github.com/kardiachain/go-kardiamain"
	"github.com/kardiachain/go-kardiamain/lib/common"
	"github.com/kardiachain/go-kardiamain/lib/p2p"
	"github.com/kardiachain/go-kardiamain/lib/rlp"
	"github.com/kardiachain/go-kardiamain/rpc"
	coreTypes "github.com/kardiachain/go-kardiamain/types"

	"github.com/kardiachain/explorer-backend/types"
)

// RPCClient return an *rpc.Client instance
type Client struct {
	clientList               map[string]*rpc.Client
	lgr                      *zap.Logger
	cache                    *redis.Client
	currentRPCUrl            string
	CurrentRequestAmount     uint64
	RequestThreshold         uint64
	CirculatingRequestAmount uint64
}

// NewKaiClient creates a client that uses the given RPC client.
func NewKaiClient(rpcUrl string, Lgr *zap.Logger) (ClientInterface, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	var clientList map[string]*rpc.Client
	for _, u := range rpcURL {
		rpcClient, err := rpc.Dial(u)
		if err != nil {
			return nil, fmt.Errorf("Failed to dial rpc %q: %v", u, err)
		}
		clientList[u] = rpcClient
	}
	if r == nil {
		db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			db = 0
		}
		r = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URI"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       db,
		})
	}
	requestThreshold, err := strconv.ParseUint(os.Getenv("REQUEST_THRESHOLD"), 10, 64)
	if err != nil {
		requestThreshold = 100000
	}
	circulatingRequestAmount, err := strconv.ParseUint(os.Getenv("CIRCULATING_REQUEST_AMOUNT"), 10, 64)
	if err != nil {
		circulatingRequestAmount = 1000
	}
	return &Client{clientList, lgr, r, rpcURL[0], 0, requestThreshold, circulatingRequestAmount}, nil
}

// TODO(trinhdn): random check to determine which node is dead
/* chooseRPCClient choose the client which has least requests sent to handle upcoming requests
 */
func (ec *Client) chooseRPCClient() error {
	ec.CirculatingRequestAmount++
	if ec.CirculatingRequestAmount < ec.RequestThreshold {
		return nil
	}
	// add circulating request amount to corresponding client
	val, err := ec.cache.Get(ec.currentRPCUrl).Uint64()
	if err != nil {
		if err == redis.Nil {
			err := ec.cache.Set(ec.currentRPCUrl, ec.CurrentRequestAmount, 0).Err()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	err = ec.cache.Set(ec.currentRPCUrl, ec.CurrentRequestAmount+val, 0).Err()
	ec.CurrentRequestAmount = 0

	// re-select another RPC client (which has minimum requests sent) to handle upcoming requests
	var minNumRequest uint64 = 0
	var nextSelectedIP string
	for ip := range ec.clientList {
		numRequest, err := ec.cache.Get(ec.currentRPCUrl).Uint64()
		if err != nil {
			if err == redis.Nil {
				err := ec.cache.Set(ec.currentRPCUrl, ec.CurrentRequestAmount, 0).Err()
				if err != nil {
					return err
				}
				nextSelectedIP = ip
				break
			}
			if minNumRequest > numRequest {
				minNumRequest = numRequest
				nextSelectedIP = ip
			}
		}
	}
	fmt.Println("nextSelectedIP", nextSelectedIP)
	ec.currentRPCUrl = nextSelectedIP

	return nil
}

// LatestBlockNumber gets latest block number
func (ec *Client) LatestBlockNumber(ctx context.Context) (uint64, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return 0, err
	}
	var result uint64
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "kai_blockNumber")
	return result, err
}

// BlockByHash returns the given full block.
//
// Use HeaderByHash if you don't need all transactions or uncle headers.
func (ec *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return ec.getBlock(ctx, "kai_getBlockByHash", hash)
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
func (ec *Client) BlockHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return ec.getBlockHeader(ctx, "kai_getBlockHeaderByHash", hash)
}

// GetTransaction returns the transaction with the given hash.
func (ec *Client) GetTransaction(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, false, err
	}
	var raw *types.Transaction
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &raw, "tx_getTransaction", hash)
	if err != nil {
		return nil, false, err
	} else if raw == nil {
		return nil, false, kardia.NotFound
	}
	return raw, raw.BlockNumber == 0, nil
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var r *types.Receipt
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &r, "tx_getTransactionReceipt", txHash.Hex())
	if err == nil {
		if r == nil {
			return nil, kardia.NotFound
		}
	}
	return r, err
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockHash common.Hash, blockNumber uint64) (string, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return "", err
	}
	var result string
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "account_balance", account, blockHash, blockNumber)
	return result, err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber uint64) ([]byte, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var result common.Bytes
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "kai_getStorageAt", account, key, blockNumber)
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account common.Address, blockNumber uint64) ([]byte, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var result common.Bytes
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "kai_getCode", account, blockNumber)
	return result, err
}

// NonceAt returns the account nonce of the given account.
func (ec *Client) NonceAt(ctx context.Context, account common.Address) (uint64, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return 0, err
	}
	var result uint64
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "account_nonce", account)
	return result, err
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the GetTransactionReceipt method to get the
// contract address after the transaction has been mined.
// TODO(trinhdn): verify which types of tx is suitable for this API
func (ec *Client) SendRawTransaction(ctx context.Context, tx *coreTypes.Transaction) error {
	err := ec.chooseRPCClient()
	if err != nil {
		return err
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.clientList[ec.currentRPCUrl].CallContext(ctx, nil, "tx_sendRawTransaction", common.ToHex(data))
}

func (ec *Client) Peers(ctx context.Context) ([]*p2p.PeerInfo, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var result []*p2p.PeerInfo
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "node_peers")
	return result, err
}

func (ec *Client) NodeInfo(ctx context.Context) (*p2p.NodeInfo, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var result *p2p.NodeInfo
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "node_nodeInfo")
	return result, err
}

func (ec *Client) Datadir(ctx context.Context) (string, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return "", err
	}
	var result string
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "node_datadir")
	return result, err
}

func (ec *Client) Validator(ctx context.Context) []map[string]interface{} {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil
	}
	var result []map[string]interface{}
	_ = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "kai_validator")
	return result
}

func (ec *Client) Validators(ctx context.Context) []map[string]interface{} {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil
	}
	var result []map[string]interface{}
	_ = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &result, "kai_validators")
	return result
}

func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var raw types.Block
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func (ec *Client) getBlockHeader(ctx context.Context, method string, args ...interface{}) (*types.Header, error) {
	err := ec.chooseRPCClient()
	if err != nil {
		return nil, err
	}
	var raw types.Header
	err = ec.clientList[ec.currentRPCUrl].CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}
