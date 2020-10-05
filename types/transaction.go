package types

import (
	"math/big"

	"github.com/kardiachain/go-kardiamain/lib/common"
	"github.com/kardiachain/go-kardiamain/types"
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

type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockHeight uint64   `json:"blockHeight"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     uint     `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	Index       uint     `json:"logIndex"`
	Removed     bool     `json:"removed"`
}

type Receipt struct {
	BlockHash         string       `json:"blockHash"`
	BlockHeight       uint64       `json:"blockHeight"`
	TransactionHash   string       `json:"transactionHash"`
	TransactionIndex  uint64       `json:"transactionIndex"`
	From              string       `json:"from"`
	To                string       `json:"to"`
	GasUsed           uint64       `json:"gasUsed"`
	CumulativeGasUsed uint64       `json:"cumulativeGasUsed"`
	ContractAddress   string       `json:"contractAddress"`
	Logs              []Log        `json:"logs"`
	LogsBloom         types.Bloom  `json:"logsBloom"`
	Root              common.Bytes `json:"root"`
	Status            uint         `json:"status"`
}

type TransactionList struct {
	Transactions []*Transaction `json:"txs"`
}
