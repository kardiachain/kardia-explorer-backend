// Package kardia
package kardia

import (
	"context"

	"github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/explorer-backend/types"
)

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
