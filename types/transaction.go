package types

import (
	"math/big"

	"github.com/kardiachain/go-kardiamain/lib/common"
)

type Transaction struct {
	TxHash           string        `json:"hash" bson:"hash"`
	To               string        `json:"to" bson:"to"`
	From             string        `json:"from" bson:"from"`
	Status           bool          `json:"status" bson:"status"`
	ContractAddress  string        `json:"contract_address" bson:"contract_address"`
	Value            string        `json:"value" bson:"value"`
	GasPrice         string        `json:"gasPrice" bson:"gasPrice"`
	GasFee           big.Int       `json:"gas" bson:"gas"`
	GasLimit         common.Uint64 `json:"gasLimit" bson:"gasLimit"`
	BlockNumber      uint64        `json:"blockNumber" bson:"blockNumber"`
	Nonce            string        `json:"nonce" bson:"nonce"`
	BlockHash        string        `json:"blockHash" bson:"blockHash"`
	CreatedAt        uint64        `json:"time" bson:"time"`
	InputData        string        `json:"input" bson:"input"`
	Logs             string        `json:"logs" bson:"logs"`
	TransactionIndex uint          `json:"transactionIndex"`

	ReceiptReceived bool `json:"-" bson:"receipt_received"`
}

type TransactionList struct {
	Transactions []*Transaction `json:"txs"`
}

type Tps struct {
	Time   uint64 `json:"time"`
	NumTxs uint64 `json:"num_txs"`
}
